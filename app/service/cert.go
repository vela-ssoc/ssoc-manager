package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"

	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/ssoc-common-mb/param/request"
	"github.com/vela-ssoc/ssoc-manager/app/internal/param"
	"github.com/vela-ssoc/ssoc-manager/errcode"
	"github.com/vela-ssoc/ssoc-manager/param/mrequest"
)

type CertService interface {
	Page(ctx context.Context, pager param.Pager) (int64, []*model.Certificate)
	Indices(ctx context.Context, idx param.Indexer) request.IDNames
	Create(ctx context.Context, dat *mrequest.CertCreate) error
	Update(ctx context.Context, dat *mrequest.CertUpdate) error
	Delete(ctx context.Context, id int64) error
}

func Cert(qry *query.Query) CertService {
	return &certService{
		qry: qry,
	}
}

type certService struct {
	qry *query.Query
}

func (biz *certService) Page(ctx context.Context, pager param.Pager) (int64, []*model.Certificate) {
	tbl := biz.qry.Certificate
	dao := tbl.WithContext(ctx).
		Order(tbl.ID)
	if kw := pager.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw))
	}
	count, _ := dao.Count()
	if count == 0 {
		return 0, nil
	}
	dats, _ := dao.Scopes(pager.Scope(count)).Find()

	return count, dats
}

func (biz *certService) Indices(ctx context.Context, idx param.Indexer) request.IDNames {
	tbl := biz.qry.Certificate
	dao := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.Name).
		Order(tbl.ID)
	if kw := idx.Keyword(); kw != "" {
		dao.Where(tbl.Name.Like(kw))
	}

	var dats request.IDNames
	_ = dao.Scopes(idx.Scope).Scan(&dats)

	return dats
}

func (biz *certService) Create(ctx context.Context, dat *mrequest.CertCreate) error {
	// 检查证书与私钥是否匹配
	pair, err := tls.X509KeyPair([]byte(dat.Certificate), []byte(dat.PrivateKey))
	if err != nil || len(pair.Certificate) == 0 {
		return errcode.ErrCertMatchKey
	}

	cert, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return errcode.ErrCertificate
	}

	ips := make([]string, 0, len(cert.IPAddresses))
	for _, a := range cert.IPAddresses {
		ips = append(ips, a.String())
	}
	uris := make([]string, 0, len(cert.URIs))
	for _, u := range cert.URIs {
		uris = append(uris, u.String())
	}

	iss, sub := cert.Issuer, cert.Subject
	insert := &model.Certificate{
		Name:           dat.Name,
		Certificate:    dat.Certificate,
		PrivateKey:     dat.PrivateKey,
		Version:        cert.Version,
		IssCountry:     iss.Country,
		IssProvince:    iss.Province,
		IssOrg:         iss.Organization,
		IssCN:          iss.CommonName,
		IssOrgUnit:     iss.OrganizationalUnit,
		SubCountry:     sub.Country,
		SubOrg:         sub.Organization,
		SubProvince:    sub.Province,
		SubCN:          sub.CommonName,
		DNSNames:       cert.DNSNames,
		IPAddresses:    ips,
		EmailAddresses: cert.EmailAddresses,
		URIs:           uris,
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
	}

	return biz.qry.Certificate.
		WithContext(ctx).
		Create(insert)
}

func (biz *certService) Update(ctx context.Context, dat *mrequest.CertUpdate) error {
	tbl := biz.qry.Certificate
	old, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.CreatedAt).
		Where(tbl.ID.Eq(dat.ID)).
		First()
	if err != nil {
		return err
	}

	// 检查证书与私钥是否匹配
	pair, err := tls.X509KeyPair([]byte(dat.Certificate), []byte(dat.PrivateKey))
	if err != nil || len(pair.Certificate) == 0 {
		return errcode.ErrCertMatchKey
	}

	cert, err := x509.ParseCertificate(pair.Certificate[0])
	if err != nil {
		return errcode.ErrCertificate
	}

	ips := make([]string, 0, len(cert.IPAddresses))
	for _, a := range cert.IPAddresses {
		ips = append(ips, a.String())
	}
	uris := make([]string, 0, len(cert.URIs))
	for _, u := range cert.URIs {
		uris = append(uris, u.String())
	}

	iss, sub := cert.Issuer, cert.Subject
	save := &model.Certificate{
		ID:             old.ID,
		Name:           dat.Name,
		Certificate:    dat.Certificate,
		PrivateKey:     dat.PrivateKey,
		Version:        cert.Version,
		IssCountry:     iss.Country,
		IssProvince:    iss.Province,
		IssOrg:         iss.Organization,
		IssCN:          iss.CommonName,
		IssOrgUnit:     iss.OrganizationalUnit,
		SubCountry:     sub.Country,
		SubOrg:         sub.Organization,
		SubProvince:    sub.Province,
		SubCN:          sub.CommonName,
		DNSNames:       cert.DNSNames,
		IPAddresses:    ips,
		EmailAddresses: cert.EmailAddresses,
		URIs:           uris,
		NotBefore:      cert.NotBefore,
		NotAfter:       cert.NotAfter,
		CreatedAt:      old.CreatedAt,
	}

	return tbl.WithContext(ctx).Save(save)
}

func (biz *certService) Delete(ctx context.Context, id int64) error {
	tbl := biz.qry.Certificate
	_, err := tbl.WithContext(ctx).
		Select(tbl.ID, tbl.CreatedAt).
		Where(tbl.ID.Eq(id)).
		First()
	if err != nil {
		return err
	}

	// 检查证书是否被使用
	brkTbl := biz.qry.Broker
	count, err := brkTbl.WithContext(ctx).Where(brkTbl.CertID.Eq(id)).Count()
	if err != nil || count != 0 {
		return errcode.ErrCertUsedByBroker
	}

	_, err = tbl.WithContext(ctx).Where(tbl.ID.Eq(id)).Delete()

	return err
}
