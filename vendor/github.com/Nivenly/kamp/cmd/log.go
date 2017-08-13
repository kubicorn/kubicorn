// Copyright © 2017 The Kamp Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

//
//import (
//	"fmt"
//	"github.com/Nivenly/kamp/local"
//	klog "github.com/Nivenly/kamp/log"
//	"github.com/spf13/cobra"
//	"log"
//	"os"
//)
//
//var logCmd = &cobra.Command{
//	Use:   "log",
//	Short: "Return cluster logs",
//	Long:  KampBannerMessage("Return log information and metrics about your kamp instance."),
//	Run: func(cmd *cobra.Command, args []string) {
//		err := RunLog(logOpt)
//		if err != nil {
//			fmt.Printf("Error: %v\n", err)
//			os.Exit(1)
//		}
//		os.Exit(0)
//	},
//}
//
//func init() {
//	RootCmd.AddCommand(logCmd)
//	logCmd.Flags().StringVarP(&logOpt.KubernetesNamespace, "namespace", "n", "default", "The Kubernetes namespace to run the container in.")
//	logCmd.SetUsageTemplate(UsageTemplate)
//}
//
//type LogOptions struct {
//	Options
//	KubernetesNamespace string
//}
//
//var logOpt = &LogOptions{}
//
//func RunLog(options *LogOptions) error {
//
//	conf, err := local.GetLocal()
//	if err != nil {
//		log.Fatal(err)
//	}
//	klog.GetLogs(conf, options.KubernetesNamespace)
//
//	return nil
//}
