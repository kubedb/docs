package server

import (
	"fmt"
	"io"
	"net"

	"github.com/appscode/kutil/tools/clientcmd"
	"github.com/kubedb/postgres/pkg/controller"
	"github.com/kubedb/postgres/pkg/server"
	"github.com/spf13/pflag"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

const defaultEtcdPathPrefix = "/registry/kubedb.com"

type PostgresServerOptions struct {
	RecommendedOptions *genericoptions.RecommendedOptions
	ExtraOptions       *ExtraOptions

	StdOut io.Writer
	StdErr io.Writer
}

func NewPostgresServerOptions(out, errOut io.Writer) *PostgresServerOptions {
	o := &PostgresServerOptions{
		// TODO we will nil out the etcd storage options.  This requires a later level of k8s.io/apiserver
		RecommendedOptions: genericoptions.NewRecommendedOptions(defaultEtcdPathPrefix, server.Codecs.LegacyCodec(admissionv1beta1.SchemeGroupVersion)),
		ExtraOptions:       NewExtraOptions(),
		StdOut:             out,
		StdErr:             errOut,
	}
	o.RecommendedOptions.Etcd = nil
	o.RecommendedOptions.Admission = nil

	return o
}

func (o PostgresServerOptions) AddFlags(fs *pflag.FlagSet) {
	o.RecommendedOptions.AddFlags(fs)
	o.ExtraOptions.AddFlags(fs)
}

func (o PostgresServerOptions) Validate(args []string) error {
	return nil
}

func (o *PostgresServerOptions) Complete() error {
	return nil
}

func (o PostgresServerOptions) Config() (*server.PostgresServerConfig, error) {
	// TODO have a "real" external address
	if err := o.RecommendedOptions.SecureServing.MaybeDefaultWithSelfSignedCerts("localhost", nil, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return nil, fmt.Errorf("error creating self-signed certificates: %v", err)
	}

	serverConfig := genericapiserver.NewRecommendedConfig(server.Codecs)
	if err := o.RecommendedOptions.ApplyTo(serverConfig, server.Scheme); err != nil {
		return nil, err
	}
	clientcmd.Fix(serverConfig.ClientConfig)

	controllerConfig := controller.NewOperatorConfig(serverConfig.ClientConfig)
	if err := o.ExtraOptions.ApplyTo(controllerConfig); err != nil {
		return nil, err
	}

	config := &server.PostgresServerConfig{
		GenericConfig:  serverConfig,
		ExtraConfig:    server.ExtraConfig{},
		OperatorConfig: controllerConfig,
	}
	return config, nil
}

func (o PostgresServerOptions) Run(stopCh <-chan struct{}) error {
	config, err := o.Config()
	if err != nil {
		return err
	}

	s, err := config.Complete().New()
	if err != nil {
		return err
	}

	return s.Run(stopCh)
}
