package service

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"log/slog"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-common-mb/datalayer/model"
	"github.com/vela-ssoc/ssoc-common-mb/datalayer/polydb"
	"github.com/vela-ssoc/ssoc-manager/appv3/request"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewCertificate(db polydb.Sourcer, log *slog.Logger) *Certificate {
	return &Certificate{
		db:  db,
		log: log,
	}
}

type Certificate struct {
	db    polydb.Sourcer
	log   *slog.Logger
	mutex sync.Mutex
}

func (crt *Certificate) List(ctx context.Context) (model.Certificates, error) {
	coll := crt.db.MDB().Certificate()

	proj := bson.M{"public_key": 0, "private_key": 0}
	opts := options.Find().SetProjection(proj)

	return coll.Find(ctx, bson.D{}, opts)
}

func (crt *Certificate) Create(ctx context.Context, param request.CertificateCreate) error {
	now := time.Now()
	pubKey, priKey := param.PublicKey, param.PrivateKey
	mod, err := crt.Parse(pubKey, priKey)
	if err != nil {
		return err
	}
	mod.UpdatedAt = now
	mod.CreatedAt = now

	coll := crt.db.MDB().Certificate()
	_, err = coll.InsertOne(ctx, mod)

	return err
}

func (crt *Certificate) Update(ctx context.Context, param request.CertificateUpdate) error {
	id := param.OID()
	now := time.Now()
	pubKey, priKey := param.PublicKey, param.PrivateKey
	mod, err := crt.Parse(pubKey, priKey)
	if err != nil {
		return err
	}
	mod.UpdatedAt = now

	coll := crt.db.MDB().Certificate()
	update := bson.M{"$set": mod}
	_, err = coll.UpdateByID(ctx, id, update)

	return nil
}

// Delete 通过 ID 删除证书。
func (crt *Certificate) Delete(ctx context.Context, ids []bson.ObjectID) error {
	if len(ids) == 0 {
		return nil
	}

	coll := crt.db.MDB().Certificate()
	filter := bson.M{"_id": bson.M{"$in": ids}}
	res, err := coll.DeleteMany(ctx, filter)
	if err != nil {
		return err
	} else if res.DeletedCount == 0 {
		return errcode.ErrNilDocument
	}

	return nil
}

func (crt *Certificate) Parse(publicKey, privateKey string) (*model.Certificate, error) {
	pubKey, priKey := []byte(publicKey), []byte(privateKey)
	cert, err := tls.X509KeyPair(pubKey, priKey)
	if err != nil {
		return nil, errcode.ErrCertificateInvalid
	}

	leaf := cert.Leaf
	sub := leaf.Subject
	ips := make([]string, 0, len(leaf.IPAddresses))
	for _, addr := range leaf.IPAddresses {
		ips = append(ips, addr.String())
	}
	uris := make([]string, 0, len(leaf.URIs))
	for _, uri := range leaf.URIs {
		uris = append(uris, uri.String())
	}

	// 计算指纹
	certSHA256, pubKeySHA256, priKeySHA256 := crt.fingerprintSHA256(cert)
	dat := &model.Certificate{
		CommonName:         sub.CommonName,
		PublicKey:          publicKey,
		PrivateKey:         privateKey,
		CertificateSHA256:  certSHA256,
		PublicKeySHA256:    pubKeySHA256,
		PrivateKeySHA256:   priKeySHA256,
		DNSNames:           leaf.DNSNames,
		IPAddresses:        ips,
		EmailAddresses:     leaf.EmailAddresses,
		URIs:               uris,
		Version:            leaf.Version,
		NotBefore:          leaf.NotBefore,
		NotAfter:           leaf.NotAfter,
		Issuer:             crt.parsePKIX(leaf.Issuer),
		Subject:            crt.parsePKIX(leaf.Subject),
		SignatureAlgorithm: cert.Leaf.SignatureAlgorithm.String(),
	}

	return dat, nil
}

func (*Certificate) parsePKIX(v pkix.Name) model.CertificatePKIXName {
	return model.CertificatePKIXName{
		Country:            v.Country,
		Organization:       v.Organization,
		OrganizationalUnit: v.OrganizationalUnit,
		Locality:           v.Locality,
		Province:           v.Province,
		StreetAddress:      v.StreetAddress,
		PostalCode:         v.PostalCode,
		SerialNumber:       v.SerialNumber,
		CommonName:         v.CommonName,
	}
}

// fingerprintSHA256 计算证书和私钥的 SHA256 指纹。
func (*Certificate) fingerprintSHA256(cert tls.Certificate) (certSHA256, pubKeySHA256, priKeySHA256 string) {
	leaf := cert.Leaf
	sum256 := sha256.Sum256(leaf.Raw)
	certSHA256 = hex.EncodeToString(sum256[:])

	if pki, _ := x509.MarshalPKIXPublicKey(leaf.PublicKey); len(pki) != 0 {
		sum := sha256.Sum256(pki)
		pubKeySHA256 = hex.EncodeToString(sum[:])
	}

	if pkcs8, _ := x509.MarshalPKCS8PrivateKey(cert.PrivateKey); len(pkcs8) != 0 {
		sum := sha256.Sum256(pkcs8)
		priKeySHA256 = hex.EncodeToString(sum[:])
	}

	return certSHA256, pubKeySHA256, priKeySHA256
}
