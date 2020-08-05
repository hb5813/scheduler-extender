package controller

import (
	"log"
	"math/rand"
	"strings"

	"k8s.io/api/core/v1"
	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

const (
	LuckyPred        = "Lucky"
	LuckyPredFailMsg = "Sorry, you're not lucky"
)

var predicatesFuncs = map[string]FitPredicate{
	LuckyPred: LuckyPredicate,
}

type FitPredicate func(pod *v1.Pod, node v1.Node) (bool, []string, error)

var predicatesSorted = []string{LuckyPred}

// filter 根据扩展程序定义的预选规则来过滤节点
// it's webhooked to pkg/scheduler/core/generic_scheduler.go#findNodesThatFit()
func filter(args schedulerapi.ExtenderArgs) *schedulerapi.ExtenderFilterResult {
	var filteredNodes []v1.Node
	failedNodes := make(schedulerapi.FailedNodesMap)
	pod := args.Pod
	// 循环遍历每个节点
	for _, node := range args.Nodes.Items {
		// 判断是否应该批准该节点
		fits, failReasons, _ := podFitsOnNode(pod, node)
		if fits { // 批准则加入filteredNodes
			filteredNodes = append(filteredNodes, node)
		} else { // 不批准则加入failedNodes
			failedNodes[node.Name] = strings.Join(failReasons, ",")
		}
	}

	result := schedulerapi.ExtenderFilterResult{
		Nodes: &v1.NodeList{
			Items: filteredNodes,
		},
		FailedNodes: failedNodes,
		Error:       "",
	}

	return &result
}

func podFitsOnNode(pod *v1.Pod, node v1.Node) (bool, []string, error) {
	fits := true
	var failReasons []string
	for _, predicateKey := range predicatesSorted {
		fit, failures, err := predicatesFuncs[predicateKey](pod, node)
		if err != nil {
			return false, nil, err
		}
		fits = fits && fit
		failReasons = append(failReasons, failures...)
	}
	return fits, failReasons, nil
}

// 累计批准的节点数
var tot_lucky_node int = 0

// 作为批准的分母的数目
var denominator int = 10

func LuckyPredicate(pod *v1.Pod, node v1.Node) (bool, []string, error) {
	// 在{0 ~ denominator}中获取一个随机数，当随机数大于累计批准的节点数时才批准
	lucky := rand.Intn(denominator) >= tot_lucky_node
	if lucky { // 如果是1，则被选中
		tot_lucky_node++
		log.Printf("pod %v/%v is lucky to fit on node %v\n", pod.Name, pod.Namespace, node.Name)
		return true, nil, nil
	}
	log.Printf("pod %v/%v is unlucky to fit on node %v\n", pod.Name, pod.Namespace, node.Name)
	return false, []string{LuckyPredFailMsg}, nil
}
