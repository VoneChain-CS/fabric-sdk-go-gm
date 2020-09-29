// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/hyperledger/fabric-sdk-go

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/cloudflare/cfssl v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.9.0
	github.com/golang/mock v1.4.3
	github.com/golang/protobuf v1.3.3
	github.com/google/certificate-transparency-go v1.1.0 // indirect
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/grantae/certinfo v0.0.0-20170412194111-59d56a35515b
	github.com/hyperledger/fabric v2.1.1+incompatible // indirect
	github.com/hyperledger/fabric-ca v1.4.8 // indirect
	github.com/hyperledger/fabric-config v0.0.5
	github.com/hyperledger/fabric-lib-go v1.0.0
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/miekg/pkcs11 v1.0.3
	github.com/mitchellh/mapstructure v1.3.3
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/procfs v0.0.5 // indirect
	github.com/spf13/cast v1.3.1
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.5.1
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	github.com/tjfoc/gmsm v1.3.2
	github.com/tjfoc/gmtls v1.2.1
	github.com/zmap/zlint v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/net v0.0.0-20200421231249-e086a090c8fd
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
	google.golang.org/genproto v0.0.0-20200423170343-7949de9c1215 // indirect
	google.golang.org/grpc v1.29.1
	gopkg.in/yaml.v2 v2.3.0
)

replace (
	github.com/Shopify/sarama => ./Shopify/sarama
	github.com/cloudflare/cfssl => ./cfssl
	github.com/spf13/cast => ./spf13/cast
	github.com/spf13/cobra => ./spf13/cobra
	github.com/spf13/jwalterweatherman => ./spf13/jwalterweatherman
	github.com/spf13/pflag => ./spf13/pflag
)

go 1.14
