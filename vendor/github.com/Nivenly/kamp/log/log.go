package log

import (
	"fmt"

	//"github.com/Nivenly/kamp/k8s"
	"github.com/Nivenly/kamp/local"

	"bufio"
	"github.com/fatih/color"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"math/rand"
	"os"
	//"sync"
)

func GetLogs(local *local.KampConfig, namespace string) error {
	//client, err := k8s.LoadClient()
	//if err != nil {
	//	return err
	//}
	//listOpts := v1.ListOptions{
	//	LabelSelector: "kamp=" + local.ProjectName,
	//}
	//pods, err := client.CoreV1().Pods(namespace).List(listOpts)
	//if err != nil {
	//	return fmt.Errorf("problem getting pods: ", err)
	//}
	//fmt.Printf("found %v pods \n", len(pods.Items))
	//
	//var wg sync.WaitGroup
	//for _, pod := range pods.Items {
	//	wg.Add(1)
	//	fmt.Printf("tailing pod: %+v \n", pod.Name)
	//	go tailPod(pod, client)
	//}
	//wg.Wait()

	return nil
}

func tailPod(p v1.Pod, cli *kubernetes.Clientset) error {
	req := cli.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &v1.PodLogOptions{Follow: true})
	fmt.Printf("request: %+v \n", req)
	readCloser, err := req.Stream()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(readCloser)
	c := AllColors[rand.Intn(len(AllColors))]
	scol := color.New(c).SprintFunc()

	for scanner.Scan() {
		fmt.Printf("%s | %s \n", scol(p.Name), scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error with the scanner in attached container", err)
	}

	return err
}
