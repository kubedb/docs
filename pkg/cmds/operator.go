/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmds

import (
	"context"

	"kubedb.dev/operator/pkg/cmds/server"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
)

func NewCmdOperator(ctx context.Context) *cobra.Command {
	o := server.NewOperatorOptions()

	cmd := &cobra.Command{
		Use:               "operator",
		Short:             "Launch KubeDB Provisioner",
		Long:              "Launch KubeDB Provisioner",
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infoln("Starting kubedb-provisioner...")

			if err := o.Complete(); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return errors.NewAggregate(err)
			}
			return o.Run(ctx)
		},
	}

	o.AddFlags(cmd.Flags())

	return cmd
}
