package main

import (
	"log"
	"net"
	"strconv"
)

type MachineManager interface {
	// Takes snapshot name, and returns the IP to connect to.
	SpinupMachine(name string) (string, error)
}

type service struct {
	name string
	/*
	 This should start at 0 and should be incremented any time a cleanup check
	 shows no connections on this port. Once it reaches 5, the port forward
	 should be deleted along with that port in the map
	*/
	connectionStatus int
}

// Listens for clients on given ports to spin up machines for the ports
type ServiceManager struct {
	machineManager MachineManager
	machineSvcCnt  map[string]int
	svcConnStatus  map[int]service
}

func NewServiceHandler(mm MachineManager) *ServiceManager {
	sh := new(ServiceManager)
	sh.machineManager = mm
	return sh
}

func (me *ServiceManager) setupService(serviceName, machineName string, port int) {
	portStr := strconv.Itoa(port)
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:"+portStr)
	if err != nil {
		log.Print("Could not resolve port and listen address:", err)
		return
	}

	// listen on port
	go func() {
		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			log.Print("Could not listen for TCP connection:", err)
		} else {
			for {
				conn, err := l.AcceptTCP()
				if err != nil {
					log.Print("Could not accept TCP connection:", err)
				} else {
					// connection accepted

					// spinup machine
					ip, err := me.machineManager.SpinupMachine(machineName)

					// setup port forwarding
					if err == nil {
						setupPortForward(ip, portStr)
						me.machineSvcCnt[machineName]++
						me.svcConnStatus[port] = service{name: serviceName, connectionStatus: 0}
					} else {
						log.Print("Could not setup machine "+machineName+":", err)
					}

					// close existing connection, not doing anything with it
					conn.Close()
				}
			}
		}
	}()
}

/*
  Periodically checks number of connections to each machine and deletes
  them when they are no longer needed
*/
func (me *ServiceManager) cleanup() {
}
