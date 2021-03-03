package repair

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"via/db"
	"via/record"
)

func EnableShutDownListener() {
	go func() {
		down := make(chan os.Signal, 1)
		signal.Notify(down, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-down
		go func() {
			tk := time.NewTicker(30*time.Second)
			<-tk.C
			log.Println("Ticker Finished")
			os.Exit(-1)
		}()
		log.Println("Preparing to close")
		log.Println("Lock Database")
		db.Lock()
		log.Println("Close Record file")
		record.CloseFile()
		log.Println("Ready to close")
		os.Exit(0)
	}()
}
