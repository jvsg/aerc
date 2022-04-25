package pgp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"git.sr.ht/~rjarry/aerc/models"
	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	"github.com/emersion/go-message/mail"
	"github.com/emersion/go-pgpmail"
	"github.com/kyoh86/xdg"
	"github.com/pkg/errors"
)

type Mail struct {
	logger *log.Logger
}

var (
	Keyring openpgp.EntityList

	locked bool
)

func (m *Mail) Init(l *log.Logger) error {
	m.logger = l
	m.logger.Println("Initializing PGP keyring")
	os.MkdirAll(path.Join(xdg.DataHome(), "aerc"), 0700)

	lockpath := path.Join(xdg.DataHome(), "aerc", "keyring.lock")
	lockfile, err := os.OpenFile(lockpath, os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		// TODO: Consider connecting to main process over IPC socket
		locked = false
	} else {
		locked = true
		lockfile.Close()
	}

	keypath := path.Join(xdg.DataHome(), "aerc", "keyring.asc")
	keyfile, err := os.Open(keypath)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		panic(err)
	}
	defer keyfile.Close()

	Keyring, err = openpgp.ReadKeyRing(keyfile)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m *Mail) Close() {
	if !locked {
		return
	}
	lockpath := path.Join(xdg.DataHome(), "aerc", "keyring.lock")
	os.Remove(lockpath)
}

func (m *Mail) getEntityByEmail(email string) (e *openpgp.Entity, err error) {
	for _, entity := range Keyring {
		ident := entity.PrimaryIdentity()
		if ident != nil && ident.UserId.Email == email {
			return entity, nil
		}
	}
	return nil, fmt.Errorf("entity not found in keyring")
}

func (m *Mail) getSignerEntityByEmail(email string) (e *openpgp.Entity, err error) {
	for _, key := range Keyring.DecryptionKeys() {
		if key.Entity == nil {
			continue
		}
		ident := key.Entity.PrimaryIdentity()
		if ident != nil && ident.UserId.Email == email {
			return key.Entity, nil
		}
	}
	return nil, fmt.Errorf("entity not found in keyring")
}

func (m *Mail) Decrypt(r io.Reader, decryptKeys openpgp.PromptFunction) (*models.MessageDetails, error) {
	md := new(models.MessageDetails)

	pgpReader, err := pgpmail.Read(r, Keyring, decryptKeys, nil)
	if err != nil {
		return nil, err
	}
	if pgpReader.MessageDetails.IsEncrypted {
		md.IsEncrypted = true
		md.DecryptedWith = pgpReader.MessageDetails.DecryptedWith.Entity.PrimaryIdentity().Name
		md.DecryptedWithKeyId = pgpReader.MessageDetails.DecryptedWith.PublicKey.KeyId
	}
	if pgpReader.MessageDetails.IsSigned {
		// we should consume the UnverifiedBody until EOF in order
		// to get the correct signature data
		data, err := ioutil.ReadAll(pgpReader.MessageDetails.UnverifiedBody)
		if err != nil {
			return nil, err
		}
		pgpReader.MessageDetails.UnverifiedBody = bytes.NewReader(data)

		md.IsSigned = true
		md.SignedBy = ""
		md.SignedByKeyId = pgpReader.MessageDetails.SignedByKeyId
		md.SignatureValidity = models.Valid
		if pgpReader.MessageDetails.SignatureError != nil {
			md.SignatureError = pgpReader.MessageDetails.SignatureError.Error()
			md.SignatureValidity = handleSignatureError(md.SignatureError)
		}
		if pgpReader.MessageDetails.SignedBy != nil {
			md.SignedBy = pgpReader.MessageDetails.SignedBy.Entity.PrimaryIdentity().Name
		}
	}
	md.Body = pgpReader.MessageDetails.UnverifiedBody
	return md, nil
}

func (m *Mail) ImportKeys(r io.Reader) error {
	keys, err := openpgp.ReadKeyRing(r)
	if err != nil {
		return err
	}
	Keyring = append(Keyring, keys...)
	if locked {
		keypath := path.Join(xdg.DataHome(), "aerc", "keyring.asc")
		keyfile, err := os.OpenFile(keypath, os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			return err
		}
		defer keyfile.Close()

		for _, key := range keys {
			if key.PrivateKey != nil {
				err = key.SerializePrivate(keyfile, &packet.Config{})
			} else {
				err = key.Serialize(keyfile)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Mail) Encrypt(buf *bytes.Buffer, rcpts []string, signerEmail string, decryptKeys openpgp.PromptFunction, header *mail.Header) (io.WriteCloser, error) {
	var err error
	var to []*openpgp.Entity
	var signer *openpgp.Entity
	if signerEmail != "" {
		signer, err = m.getSigner(signerEmail, decryptKeys)
		if err != nil {
			return nil, err
		}
	}

	for _, rcpt := range rcpts {
		toEntity, err := m.getEntityByEmail(rcpt)
		if err != nil {
			return nil, errors.Wrap(err, "no key for "+rcpt)
		}
		to = append(to, toEntity)
	}

	cleartext, err := pgpmail.Encrypt(buf, header.Header.Header,
		to, signer, nil)
	if err != nil {
		return nil, err
	}
	return cleartext, nil
}

func (m *Mail) Sign(buf *bytes.Buffer, signerEmail string, decryptKeys openpgp.PromptFunction, header *mail.Header) (io.WriteCloser, error) {
	var err error
	var signer *openpgp.Entity
	if signerEmail != "" {
		signer, err = m.getSigner(signerEmail, decryptKeys)
		if err != nil {
			return nil, err
		}
	}
	cleartext, err := pgpmail.Sign(buf, header.Header.Header, signer, nil)
	if err != nil {
		return nil, err
	}
	return cleartext, nil
}

func (m *Mail) getSigner(signerEmail string, decryptKeys openpgp.PromptFunction) (signer *openpgp.Entity, err error) {
	if err != nil {
		return nil, err
	}
	signer, err = m.getSignerEntityByEmail(signerEmail)
	if err != nil {
		return nil, err
	}

	key, ok := signer.SigningKey(time.Now())
	if !ok {
		return nil, fmt.Errorf("no signing key found for %s", signerEmail)
	}

	if !key.PrivateKey.Encrypted {
		return signer, nil
	}

	_, err = decryptKeys([]openpgp.Key{key}, false)
	if err != nil {
		return nil, err
	}

	return signer, nil
}

func handleSignatureError(e string) models.SignatureValidity {
	if e == "openpgp: signature made by unknown entity" {
		return models.UnknownEntity
	}
	if strings.HasPrefix(e, "pgpmail: unsupported micalg") {
		return models.UnsupportedMicalg
	}
	if strings.HasPrefix(e, "pgpmail") {
		return models.InvalidSignature
	}
	return models.UnknownValidity
}
