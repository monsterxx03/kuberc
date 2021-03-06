/*
Copyright © 2020 Will Xu <xyj.asmy@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/monsterxx03/kuberc/pkg/redis"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"sort"
)

var cfgFile string
var namespace string
var containerName string
var redisPort int
var restcfg *restclient.Config
var clientset *kubernetes.Clientset

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "rc",
	Short: "Manage redis cluster on k8s",
	Long: `Used as a kubectl plugin, to get redis cluster info, 
replace redis nodes on k8s.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if os.Getenv("KUBECONFIG") != "" {
			cfgFile = os.Getenv("KUBECONFIG")
		} else {
			cfgFile, err = homedir.Expand(cfgFile)
			if err != nil {
				return err
			}
		}
		restcfg, err = clientcmd.BuildConfigFromFlags("", cfgFile)
		if err != nil {
			return err
		}
		clientset, err = kubernetes.NewForConfig(restcfg)
		if err != nil {
			return err
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&redisPort, "port", "p", 6379, "redis port")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "namespace")
	rootCmd.PersistentFlags().StringVarP(&containerName, "container", "c", "", "container name")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "kubeconfig used for kubectl, will try to load from $KUBECONFIG first")
}

func getClusterPods(podname string, all bool) ([]*redis.RedisPod, error) {
	pod, err := redis.NewRedisPod(podname, containerName, namespace, redisPort, clientset, restcfg)
	if err != nil {
		return nil, err
	}
	pods := make([]*redis.RedisPod, 0)
	if all {
		if nodes, err := pod.ClusterNodes(); err != nil {
			return nil, err
		} else {
			for _, n := range nodes {
				pods = append(pods, redis.NewRedisPodWithPod(n.Pod, containerName, redisPort, clientset, restcfg))
			}
		}
	} else {
		pods = append(pods, pod)
	}
	sort.Slice(pods, func(i, j int) bool {
		return pods[i].GetName() < pods[j].GetName()
	})
	return pods, nil
}

func main() {
	Execute()
}
