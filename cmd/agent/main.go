package main

import (
	"time"

	"github.com/Vitaly898/metricsCollector/internal/agent"
)


func main(){
	collector := agent.NewCollector()
	sender := agent.NewSender("http://localhost:8080")
	a := agent.NewAgent(collector,sender,2*time.Second,10*time.Second)
	a.Run()

}