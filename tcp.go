package pipgo

import (
	"sync"
	"time"

	"github.com/plumk97/pip-go/lib/queue"
	"github.com/plumk97/pip-go/types"
)

const (
	_TCP_WIND = 65535
)

type TCPStatus int

const (
	TCPStatusNone             TCPStatus = 0
	TCPStatusWaitEstablishing TCPStatus = 1 // 已接受syn 等待回复
	TCPStatusEstablishing     TCPStatus = 2 // 已回复syn 等待确认
	TCPStatusEstablished      TCPStatus = 3
	TCPStatusFinWait1         TCPStatus = 4
	TCPStatusFinWait2         TCPStatus = 5
	TCPStatusCloseWait        TCPStatus = 6
	TCPStatusReleased         TCPStatus = 7
)

type TCP struct {
	locker sync.Mutex
	netif  *Netif

	// 连接标识
	key TCPKey

	// 包队列
	packetQueue *queue.Queue[*TCPPacket]

	// 当前链接状态
	status TCPStatus

	// ip信息
	ipHeader *types.IPHeader

	// 主动关闭时间 定期检查 防止客户端不响应ACK 导致资源占用
	finTime time.Time

	// 当前是否等待确认PUSH包
	isWaitPushAck bool

	// 源端口
	srcPort uint16

	// 目标端口
	dstPort uint16

	// 当前发送序号
	seq uint32

	// 对方当前的seq
	oppSeq uint32

	// 当前回复对方的ack
	ack uint32

	// mss
	mss uint16

	// 对方的mss
	oppMss uint16

	// 接收窗口大小
	wind uint16

	// 对方的窗口大小
	oppWind uint32

	// 窗口缩放位移位数
	oppWindShift uint8

	// 事件
	events []any

	// 外部使用-用于区分
	Arg any

	// 建立连接完成回调
	OnConnected func(tcp *TCP)

	// 关闭回调 在这个时候资源已经释放完成
	OnClosed func(tcp *TCP, arg any)

	// 数据接收回调
	// 数据处理完成需要调用Received更新窗口
	OnReceived func(tcp *TCP, data []byte)

	// 数据发送完成回调 writeen_len完成发送的字节
	// @param writeen_len 已经发送的字节长度 如果为0 则代表之前对方的wind为0 当前已经更新 可以继续写入
	// @param has_push 是否包含push包
	OnWritten func(tcp *TCP, writtenLen int, hasPush bool)
}

func newTCP(netif *Netif) *TCP {

	tcp := &TCP{
		netif:       netif,
		packetQueue: queue.NewQueue[*TCPPacket](),
		mss:         MTU - 40,
		wind:        _TCP_WIND,
	}
	return tcp

}
