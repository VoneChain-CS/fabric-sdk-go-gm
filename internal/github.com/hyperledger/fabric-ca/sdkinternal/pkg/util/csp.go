/*
Copyright IBM Corp. 2017 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
/*
Notice: This file has been modified for Hyperledger Fabric SDK Go usage.
Please review third_party pinning scripts and patches for more details.
*/

package util

import (
	"crypto"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/gm"
	cspsigner "github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/signer"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/tjfoc/gmsm/sm2"
	tls "github.com/tjfoc/gmtls"
	"io/ioutil"
	"strings"

	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/core"

	"github.com/VoneChain-CS/fabric-sdk-go-gm/cfssl/csr"
	factory "github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric-ca/sdkpatch/cryptosuitebridge"
	log "github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric-ca/sdkpatch/logbridge"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/pkg/errors"
)

// getBCCSPKeyOpts generates a key as specified in the request.
// This supports ECDSA.
func getBCCSPKeyOpts(kr *csr.KeyRequest, ephemeral bool) (opts core.KeyGenOpts, err error) {
	if kr == nil {
		return factory.GetECDSAKeyGenOpts(ephemeral), nil
	}
	log.Debugf("generate key from request: algo=%s, size=%d", kr.Algo(), kr.Size())
	switch kr.Algo() {
	case "ecdsa":
		switch kr.Size() {
		case 256:
			return factory.GetECDSAP256KeyGenOpts(ephemeral), nil
		case 384:
			return factory.GetECDSAP384KeyGenOpts(ephemeral), nil
		case 521:
			// Need to add curve P521 to bccsp
			// return &bccsp.ECDSAP512KeyGenOpts{Temporary: false}, nil
			return nil, errors.New("Unsupported ECDSA key size: 521")
		default:
			return nil, errors.Errorf("Invalid ECDSA key size: %d", kr.Size())
		}
	case "gmsm2":
		return &bccsp.GMSM2KeyGenOpts{Temporary: ephemeral}, nil
	default:
		return nil, errors.Errorf("Invalid algorithm: %s", kr.Algo())
	}
}

// GetSignerFromCert load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCert(cert *x509.Certificate, csp core.CryptoSuite) (core.Key, crypto.Signer, error) {
	if csp == nil {
		return nil, nil, errors.New("CSP was not initialized")
	}
	// get the public key in the right format
	certPubK, err := csp.KeyImport(cert, factory.GetX509PublicKeyImportOpts(true))
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
	}
	// Get the key given the SKI value
	ski := certPubK.SKI()
	privateKey, err := csp.GetKey(ski)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Could not find matching private key for SKI")
	}
	// BCCSP returns a public key if the private key for the SKI wasn't found, so
	// we need to return an error in that case.
	if !privateKey.Private() {
		return nil, nil, errors.Errorf("The private key associated with the certificate with SKI '%s' was not found", hex.EncodeToString(ski))
	}
	// Construct and initialize the signer
	signer, err := factory.NewCspSigner(csp, privateKey)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed to load ski from bccsp")
	}
	return privateKey, signer, nil
}

// GetSignerFromCertFile load skiFile and load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCertFile(certFile string, csp core.CryptoSuite) (core.Key, crypto.Signer, *x509.Certificate, error) {
	// Load cert file
	certBytes, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "Could not read certFile '%s'", certFile)
	}
	// Parse certificate
	parsedCa, err := helpers.ParseCertificatePEM(certBytes)
	if err != nil {
		return nil, nil, nil, err
	}
	// Get the signer from the cert
	key, cspSigner, err := GetSignerFromCert(parsedCa, csp)
	return key, cspSigner, parsedCa, err
}

// BCCSPKeyRequestGenerate generates keys through BCCSP
// somewhat mirroring to cfssl/req.KeyRequest.Generate()
func BCCSPKeyRequestGenerate(req *csr.CertificateRequest, myCSP core.CryptoSuite) (core.Key, crypto.Signer, error) {
	log.Infof("generating key: %+v", req.KeyRequest)
	keyOpts, err := getBCCSPKeyOpts(req.KeyRequest, false)
	if err != nil {
		return nil, nil, err
	}
	key, err := myCSP.KeyGen(keyOpts)
	if err != nil {
		return nil, nil, err
	}
	cspSigner, err := factory.NewCspSigner(myCSP, key)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed initializing CryptoSigner")
	}
	return key, cspSigner, nil
}

// ImportBCCSPKeyFromPEM attempts to create a private BCCSP key from a pem file keyFile
func ImportBCCSPKeyFromPEM(keyFile string, myCSP core.CryptoSuite, temporary bool) (core.Key, error) {
	keyBuff, err := ioutil.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	key, err := ImportBCCSPKeyFromPEMBytes(keyBuff, myCSP, temporary)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed parsing private key from key file %s", keyFile))
	}
	return key, nil
}

// ImportBCCSPKeyFromPEMBytes attempts to create a private BCCSP key from a pem byte slice
func ImportBCCSPKeyFromPEMBytes(keyBuff []byte, myCSP core.CryptoSuite, temporary bool) (core.Key, error) {
	keyFile := "pem bytes"

	key, err := sm2.ReadPrivateKeyFromMem(keyBuff, nil)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed parsing private key from %s", keyFile))
	}

	priv, err := sm2.MarshalSm2UnecryptedPrivateKey(key)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed to convert ECDSA private key for '%s'", keyFile))
	}
	sk, err := myCSP.KeyImport(priv, factory.GetGMSM2PrivateKeyImportOpts(temporary))
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("Failed to import ECDSA private key for '%s'", keyFile))
	}
	return sk, nil

}

// LoadX509KeyPair reads and parses a public/private key pair from a pair
// of files. The files must contain PEM encoded data. The certificate file
// may contain intermediate certificates following the leaf certificate to
// form a certificate chain. On successful return, Certificate.Leaf will
// be nil because the parsed form of the certificate is not retained.
//
// This function originated from crypto/tls/tls.go and was adapted to use a
// BCCSP Signer
func LoadX509KeyPair(certFile, keyFile []byte, csp core.CryptoSuite) (*tls.Certificate, error) {

	certPEMBlock := certFile

	cert := &tls.Certificate{}
	var skippedBlockTypes []string
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		if len(skippedBlockTypes) == 0 {
			return nil, errors.New("Failed to find PEM block in bytes")
		}
		if len(skippedBlockTypes) == 1 && strings.HasSuffix(skippedBlockTypes[0], "PRIVATE KEY") {
			return nil, errors.New("Failed to find certificate PEM data in bytes, but did find a private key; PEM inputs may have been switched")
		}
		return nil, errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)
	}

	sm2Cert, err := sm2.ParseCertificate(cert.Certificate[0])
	x509Cert := gm.ParseSm2Certificate2X509(sm2Cert)

	if err != nil {
		return nil, err
	}
	w, _ := csp.(*wrapper.CryptoSuite)

	_, cert.PrivateKey, err = GetSignerFromCert2(x509Cert, w.BCCSP)
	if err != nil {
		if keyFile != nil {
			log.Debugf("Could not load TLS certificate with BCCSP: %s", err)
			log.Debug("Attempting fallback with provided certfile and keyfile")
			fallbackCerts, err := tls.X509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, errors.Wrap(err, "Could not get the private key that matches the provided cert")
			}
			cert = &fallbackCerts
		} else {
			return nil, errors.WithMessage(err, "Could not load TLS certificate with BCCSP")
		}

	}

	return cert, nil
}

// GetSignerFromCert load private key represented by ski and return bccsp signer that conforms to crypto.Signer
func GetSignerFromCert2(cert *x509.Certificate, csp bccsp.BCCSP) (bccsp.Key, crypto.Signer, error) {
	if csp == nil {
		return nil, nil, errors.New("CSP was not initialized")
	}
	log.Infof("xxxx begin csp.KeyImport,cert.PublicKey is %T   csp:%T", cert.PublicKey, csp)
	switch cert.PublicKey.(type) {
	case sm2.PublicKey:
		log.Infof("xxxxx cert is sm2 puk")
	default:
		log.Infof("xxxxx cert is default puk")
	}

	sm2cert := gm.ParseX509Certificate2Sm2(cert)
	// get the public key in the right format
	certPubK, err := csp.KeyImport(sm2cert, &bccsp.X509PublicKeyImportOpts{Temporary: true})
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Failed to import certificate's public key")
	}
	kname := hex.EncodeToString(certPubK.SKI())
	log.Infof("xxxx begin csp.GetKey kname:%s", kname)
	// Get the key given the SKI value
	privateKey, err := csp.GetKey(certPubK.SKI())
	if err != nil {
		return nil, nil, fmt.Errorf("Could not find matching private key for SKI: %s", err.Error())
	}
	log.Info("xxxx begin cspsigner.New()")
	// Construct and initialize the signer
	signer, err := cspsigner.New(wrapper.NewCryptoSuite(csp), wrapper.GetKey(privateKey))
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to load ski from bccsp: %s", err.Error())
	}
	log.Info("xxxx end GetSignerFromCert successfuul")
	return privateKey, signer, nil
}

func LoadX509KeyPairSM2(certFile, keyFile string, csp bccsp.BCCSP) (*tls.Certificate, error) {

	certPEMBlock, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	cert := &tls.Certificate{}
	var skippedBlockTypes []string
	for {
		var certDERBlock *pem.Block
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		} else {
			skippedBlockTypes = append(skippedBlockTypes, certDERBlock.Type)
		}
	}

	if len(cert.Certificate) == 0 {
		if len(skippedBlockTypes) == 0 {
			return nil, errors.Errorf("Failed to find PEM block in file %s", certFile)
		}
		if len(skippedBlockTypes) == 1 && strings.HasSuffix(skippedBlockTypes[0], "PRIVATE KEY") {
			return nil, errors.Errorf("Failed to find certificate PEM data in file %s, but did find a private key; PEM inputs may have been switched", certFile)
		}
		return nil, errors.Errorf("Failed to find \"CERTIFICATE\" PEM block in file %s after skipping PEM blocks of the following types: %v", certFile, skippedBlockTypes)
	}

	sm2Cert, err := sm2.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}

	x509Cert := gm.ParseSm2Certificate2X509(sm2Cert)
	_, cert.PrivateKey, err = GetSignerFromCert(x509Cert, wrapper.NewCryptoSuite(csp))
	if err != nil {
		if keyFile != "" {
			log.Debugf("Could not load TLS certificate with BCCSP: %s", err)
			log.Debugf("Attempting fallback with certfile %s and keyfile %s", certFile, keyFile)
			fallbackCerts, err := tls.LoadX509KeyPair(certFile, keyFile)
			if err != nil {
				return nil, errors.Wrapf(err, "Could not get the private key %s that matches %s", keyFile, certFile)
			}
			cert = &fallbackCerts
		} else {
			return nil, errors.WithMessage(err, "Could not load TLS certificate with BCCSP")
		}

	}

	return cert, nil
}
