/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sw

import (
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp"
	//bccspSw "github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/factory/sw"
	bccspGm "github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/factory"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/internal/github.com/hyperledger/fabric/bccsp/sw"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/logging"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/common/providers/core"
	"github.com/VoneChain-CS/fabric-sdk-go-gm/pkg/core/cryptosuite/bccsp/wrapper"
	"github.com/pkg/errors"
)

var logger = logging.NewLogger("fabsdk/core")

//GetSuiteByConfig returns cryptosuite adaptor for bccsp loaded according to given config
func GetSuiteByConfig(config core.CryptoSuiteConfig) (core.CryptoSuite, error) {
	// TODO: delete this check?
	if config.SecurityProvider() != "gm" {
		return nil, errors.Errorf("Unsupported BCCSP Provider: %s", config.SecurityProvider())
	}

	opts := getOptsByConfig(config)
	bccsp, err := getBCCSPFromOpts(opts)
	if err != nil {
		return nil, err
	}
	return wrapper.NewCryptoSuite(bccsp), nil
}

//GetSuiteWithDefaultEphemeral returns cryptosuite adaptor for bccsp with default ephemeral options (intended to aid testing)
func GetSuiteWithDefaultEphemeral() (core.CryptoSuite, error) {
	opts := getEphemeralOpts()

	bccsp, err := getBCCSPFromOpts(opts)
	if err != nil {
		return nil, err
	}
	return wrapper.NewCryptoSuite(bccsp), nil
}

func getBCCSPFromOpts(config *bccspGm.SwOpts) (bccsp.BCCSP, error) {
	f := &bccspGm.GMFactory{}

	config2 := &bccspGm.FactoryOpts{}
	config2.SwOpts = config
	config2.ProviderName = "GM"
	csp, err := f.Get(config2)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not initialize BCCSP %s", f.Name())
	}
	return csp, nil
}

// GetSuite returns a new instance of the software-based BCCSP
// set at the passed security level, hash family and KeyStore.
func GetSuite(securityLevel int, hashFamily string, keyStore bccsp.KeyStore) (core.CryptoSuite, error) {
	bccsp, err := sw.NewWithParams(securityLevel, hashFamily, keyStore)
	if err != nil {
		return nil, err
	}
	return wrapper.NewCryptoSuite(bccsp), nil
}

//GetOptsByConfig Returns Factory opts for given SDK config
func getOptsByConfig(c core.CryptoSuiteConfig) *bccspGm.SwOpts {
	opts := &bccspGm.SwOpts{
		HashFamily: c.SecurityAlgorithm(),
		SecLevel:   c.SecurityLevel(),
		FileKeystore: &bccspGm.FileKeystoreOpts{
			KeyStorePath: c.KeyStorePath(),
		},
	}
	logger.Debug("Initialized GM cryptosuite")

	return opts
}

func getEphemeralOpts() *bccspGm.SwOpts {
	opts := &bccspGm.SwOpts{
		HashFamily: "GMSM2",
		SecLevel:   256,
		Ephemeral:  false,
	}
	logger.Debug("Initialized ephemeral SW cryptosuite with default opts")

	return opts
}
