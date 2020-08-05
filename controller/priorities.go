package controller

import (
	"log"
	"math/rand"

	schedulerapi "k8s.io/kubernetes/pkg/scheduler/api"
)

// It'd better to only define one custom priority per extender
// as current extender interface only supports one single weight mapped to one extender
// and also it returns HostPriorityList, rather than []HostPriorityList

const (
	// lucky priority gives a random [0, schedulerapi.MaxPriority] score
	// currently schedulerapi.MaxPriority is 10
	luckyPrioMsg = "pod %v/%v is lucky to get score %v\n"
)

// it's webhooked to pkg/scheduler/core/generic_scheduler.go#PrioritizeNodes()
// you can't see existing scores calculated so far by default scheduler
// instead, scores output by this function will be added back to default scheduler
func prioritize(args schedulerapi.ExtenderArgs) *schedulerapi.HostPriorityList {
	pod := args.Pod
	nodes := args.Nodes.Items

	hostPriorityList := make(schedulerapi.HostPriorityList, len(nodes))
	for i, node := range nodes {
		// 取可申请的CPU和内存比例
		cpu := node.Status.Allocatable.Cpu().MilliValue()
		memory := node.Status.Allocatable.Memory().MilliValue()
		// 因为上面取得的值是瞬时值，完整的CPU可使用率的计算应加入滑动窗口的操作
		// 所以这里还应保留随机化的元素，即原来的随机打分情况
		// 打分算法为：CPU和内存的权重各占50%，二者取平均得到资源的平均可使用率，然后乘上原有的随机化打分
		score := int((cpu + memory) / 2)) * rand.Intn(schedulerapi.MaxPriority + 1)
		// score := rand.Intn(schedulerapi.MaxPriority + 1)
		log.Printf(luckyPrioMsg, pod.Name, pod.Namespace, score)
		hostPriorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: score,
		}
	}

	return &hostPriorityList
}

