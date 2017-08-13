package kubernetes

import (
	"github.com/Nivenly/kamp/runner"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"sync"
)

type Runner struct {
	Options *Options
}

type Options struct {
	runner.Options
	Namespace string
}

func NewKubernetesRunner(options *Options) runner.Runner {
	return &Runner{
		Options: options,
	}
}


// Run is the procederual logic for running an arbitrary container in Kubernetes with a locally mounted volume.
func (r *Runner) Run() error {
	wg := sync.WaitGroup{}
	errchan := make(chan error)

	go r.RunLocalSSHServer(wg, errchan)
	wg.Add(1)

	err := r.AttachToPod()
	if err != nil {
		return err
	}
	waitchan := make(chan int)
	go func() {
		wg.Wait()
	}()

	// Hang until we either error or complete
	select {
	case <-waitchan:
		return nil
	case err := <-errchan:
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Runner) RunLocalSSHServer(wg sync.WaitGroup, errchan chan error) {
	errchan <- nil
	wg.Done()
}

func (r *Runner) AttachToPod() error {
	f := cmdutil.NewFactory(nil)
	kubectlcmd := cmd.NewKubectlCommand(f, os.Stdin, os.Stdout, os.Stderr)
	runcmd := cmd.NewCmdRun(f, os.Stdin, os.Stdout, os.Stderr)
	attachcmd := cmd.NewCmdAttach(f, os.Stdin, os.Stdout, os.Stderr)
	kubectlcmd.AddCommand(runcmd)
	kubectlcmd.AddCommand(attachcmd)

	// kubectl run flag

	// --image
	runcmd.Flags().Set("image", r.Options.ImageQuery)

	// -i
	runcmd.Flags().Set("stdin", "1")

	// -t
	runcmd.Flags().Set("tty", "1")

	// --restart
	runcmd.Flags().Set("restart", "Never")

	// --rm
	runcmd.Flags().Set("rm", "1")

	// --attach
	runcmd.Flags().Set("attach", "1")

	// Spoof args
	args := []string{r.Options.Name}

	if err := cmd.Run(f, os.Stdin, os.Stdout, os.Stderr, runcmd, args, 1); err != nil {
		return err
	}
	return nil
}
