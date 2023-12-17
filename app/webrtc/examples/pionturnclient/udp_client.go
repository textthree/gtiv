package pionturnclient

import (
	"fmt"
	"github.com/text3cn/goodle/providers/goodlog"
	"net"

	"time"

	"github.com/pion/logging"
	"github.com/pion/turn/v2"
)

func PionUdpClient() {

	host := "192.168.1.200"
	//host := "192.168.1.200"
	//host := "118.25.6.217"
	port := 5800
	user := "your_user"
	pwd := "your_pwd"
	realm := "gtiv"

	conn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		goodlog.Error(err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			goodlog.Error(closeErr)
		}
	}()

	turnServerAddr := fmt.Sprintf("%s:%d", host, port)

	cfg := &turn.ClientConfig{
		STUNServerAddr: turnServerAddr,
		TURNServerAddr: turnServerAddr,
		Conn:           conn,
		Username:       user,
		Password:       pwd,
		Realm:          realm,
		LoggerFactory:  logging.NewDefaultLoggerFactory(),
	}

	client, err := turn.NewClient(cfg)
	if err != nil {
		goodlog.Error(err)
	}
	defer client.Close()

	// 开始监听连接
	err = client.Listen()
	if err != nil {
		goodlog.Error(err)
	}

	// 在 turn 服务器上分配中继 socket，如果成功则返回对方的 socket
	relayConn, err := client.Allocate()
	if err != nil {
		goodlog.Error(err)
	}
	defer func() {
		if closeErr := relayConn.Close(); closeErr != nil {
			goodlog.Error(closeErr)
		}
	}()

	// relayConn的本地地址实际上是传输 TURN 服务器上分配的地址。
	goodlog.Green("relayed-address=" + relayConn.LocalAddr().String())

	// ping 测试
	err = doPingTest(client, relayConn)
	if err != nil {
		goodlog.Error("ping 失败", err)
	}
}

// ping 中继服务器
func doPingTest(client *turn.Client, relayConn net.PacketConn) error {
	// Send BindingRequest to learn our external IP
	mappedAddr, err := client.SendBindingRequest()
	if err != nil {
		goodlog.Error(err)
		return err
	}

	// Set up pinger socket (pingerConn)
	pingerConn, err := net.ListenPacket("udp4", "0.0.0.0:0")
	if err != nil {
		goodlog.Error(err)
	}
	defer func() {
		if closeErr := pingerConn.Close(); closeErr != nil {
			panic(closeErr)
		}
	}()

	// Punch a UDP hole for the relayConn by sending a data to the mappedAddr.
	// This will trigger a TURN client to generate a permission request to the
	// TURN server. After this, packets from the IP address will be accepted by
	// the TURN server.
	_, err = relayConn.WriteTo([]byte("Hello"), mappedAddr)
	if err != nil {
		goodlog.Error(err)
		return err
	}

	// Start read-loop on pingerConn
	go func() {
		buf := make([]byte, 1600)
		for {
			n, from, pingerErr := pingerConn.ReadFrom(buf)
			if pingerErr != nil {
				goodlog.Pink(pingerErr)
				break
			}
			msg := string(buf[:n])
			if sentAt, pingerErr := time.Parse(time.RFC3339Nano, msg); pingerErr == nil {
				rtt := time.Since(sentAt)
				goodlog.Errorf("%d bytes from from %s time=%d ms\n", n, from.String(), int(rtt.Seconds()*1000))
			}
		}
	}()

	// Start read-loop on relayConn
	go func() {
		buf := make([]byte, 1600)
		for {
			n, from, readerErr := relayConn.ReadFrom(buf)
			if readerErr != nil {
				break
			}

			// Echo back
			if _, readerErr = relayConn.WriteTo(buf[:n], from); readerErr != nil {
				break
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)

	// Send 10 packets from relayConn to the echo server
	for i := 0; i < 10; i++ {
		msg := time.Now().Format(time.RFC3339Nano)
		_, err = pingerConn.WriteTo([]byte(msg), relayConn.LocalAddr())
		if err != nil {
			goodlog.Error(err)
			return err
		}
		// For simplicity, this example does not wait for the pong (reply).
		// Instead, sleep 1 second.
		time.Sleep(time.Second)
	}

	return nil
}
