# 实验五：容器管理-k8s搭建与调度器定制

## scheduler-extender的工作逻辑

### 过滤部分

该部分在原有的完全随机的策略下引入了模拟退火思想，使得当已批准的节点越多时，被批准的概率越小。
在{0 ~ denominator}中获取一个随机数，当随机数大于累计批准的节点数时才批准。

```go
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
```

### 打分部分

该部分在原有的完全随机的策略下引入了评估当前节点上剩余资源的机制。
先取可申请的CPU和内存比例，CPU和内存的权重各占50%，而后二者取平均得到资源的平均可使用率。因为上面取得的值是瞬时值，完整的CPU可使用率的计算应加入滑动窗口的操作，所以这里还应保留随机化的元素，即原来的随机打分情况。
故打分算法为：资源的平均可使用率乘以原有的随机化打分。

```go
func prioritize(args schedulerapi.ExtenderArgs) *schedulerapi.HostPriorityList {
	pod := args.Pod
	nodes := args.Nodes.Items

	hostPriorityList := make(schedulerapi.HostPriorityList, len(nodes))
	for i, node := range nodes {
		// 取可申请的CPU和内存比例
		cpu := node.Status.Allocatable.Cpu().MilliValue()
		memory := node.Status.Allocatable.Memory().MilliValue()
		// 打分算法为：CPU和内存的权重各占50%，二者取平均得到资源的平均可使用率，然后乘上原有的随机化打分
		score := int((cpu + memory) / 2 * rand.Intn(schedulerapi.MaxPriority + 1))
		// score := rand.Intn(schedulerapi.MaxPriority + 1)
		log.Printf(luckyPrioMsg, pod.Name, pod.Namespace, score)
		hostPriorityList[i] = schedulerapi.HostPriority{
			Host:  node.Name,
			Score: score,
		}
	}

	return &hostPriorityList
}
```