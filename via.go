package main

import (
	"fmt"
	"github.com/AkvicorEdwards/arg"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"via/db"
	"via/def"
	"via/handler"
	"via/record"
	"via/repair"
)

func main() {
	EnableOption()
	record.Enable()
	repair.EnableShutDownListener()
	repair.CheckPath()
	handler.ParsePrefix()
	addr := fmt.Sprintf("%s:%d", def.ADDR, def.PORT)
	server := http.Server{
		Addr:        addr,
		Handler:     &handler.MyHandler{},
		ReadTimeout: 24 * time.Hour,
	}
	log.Println(addr)
	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func EnableOption() {
	arg.RootCommand.Size = 0
	arg.RootCommand.Executor = func([]string) error { return nil }
	option("init", 0, func(str []string) error {
		repair.InitDatabase()
		return nil
	})
	option("-db", 1, func(str []string) error {
		def.DBFilename = str[1]
		return nil
	})
	option("-port", 1, func(str []string) error {
		var err error = nil
		def.PORT, err = strconv.Atoi(str[1])
		if err != nil || def.PORT <= 0 || def.PORT > 65535 {
			log.Println("PORT ERROR")
			os.Exit(-1)
		}
		return nil
	})
	option("-sd", 1, func(str []string) error {
		def.SessionDomain = str[1]
		return nil
	})
	option("-sp", 1, func(str []string) error {
		def.SessionPath = str[1]
		return nil
	})
	option("-sn", 1, func(str []string) error {
		def.SessionName = str[1]
		return nil
	})
	option("-u", 1, func(str []string) error {
		def.Username = str[1]
		return nil
	})
	option("-p", 1, func(str []string) error {
		def.Password = str[1]
		return nil
	})
	option("-path", 1, func(str []string) error {
		def.Path = str[1]
		return nil
	})
	option("-button-delete", 0, func(str []string) error {
		def.DeleteButton = false
		return nil
	})
	option("-record-off", 0, func(str []string) error {
		def.RecordEnable = false
		return nil
	})
	option("-record-file-off", 0, func(str []string) error {
		def.RecordEnableFile = false
		return nil
	})
	m5 := false
	s6 := false
	cl := false
	n := 4
	command("verify", 0, func([]string) error {
		if !m5 && !s6 {
			m5 = true
			s6 = true
		}
		// File
		res := db.Verify(m5, s6, cl, n)
		fmt.Println("Illegal Path", res.Path.Illegal)
		fmt.Println("Illegal File", res.File.Illegal)
		fmt.Println("Damaged Path", res.Path.Damaged)
		fmt.Println("Damaged File", res.File.Damaged)
		fmt.Println("Missing Path", res.Path.Missing)
		fmt.Println("Missing File", res.File.Missing)
		fmt.Println("Deleted Path", res.Path.Deleted.Val())
		fmt.Println("Deleted File", res.File.Deleted.Val())
		fmt.Println("Error Path", res.Path.Error)
		fmt.Println("Error File", res.File.Error)
		fmt.Println("PassedMD5 File", res.File.PassedMD5)
		fmt.Println("PassedSHA256 File", res.File.PassedSHA256)
		// Path
		pth := db.VerifyPath()
		if pth != nil {
			fmt.Println("Path Relation:")
			for _, v := range *pth {
				fmt.Printf("Path: [%d][%s] has not established a connection with ROOT\n", v.Pid, v.Title)
			}
		} else {
			fmt.Println("Path Relation: Pass")
		}
		// Relation
		rel := db.VerifyRelation()
		if rel != nil {
			fmt.Println("File Relation:")
			for _, v := range *rel {
				fmt.Printf("File: [%d][%s] has not established a connection with Dir\n", v.Fid, v.Title)
			}
		} else {
			fmt.Println("File Relation: Pass")
		}
		os.Exit(0)
		return nil
	})
	options("verify", "-m", 0, func(str []string) error {
		m5 = true
		return nil
	})
	options("verify", "-s", 0, func(str []string) error {
		s6 = true
		return nil
	})
	options("verify", "-c", 0, func(str []string) error {
		cl = true
		return nil
	})
	options("verify", "-n", 1, func(str []string) error {
		tn, err := strconv.Atoi(str[1])
		if err == nil {
			n = tn
		}
		return nil
	})

	arg.EnableOptionCombination()

	wrap(arg.Parse())
}

func option(opt string, size int, f arg.FuncExecutor) {
	wrap(arg.AddOption([]string{opt}, 0, size, 0, "",
		"", "", "", f, nil))
}

func options(opt1, opt2 string, size int, f arg.FuncExecutor) {
	wrap(arg.AddOption([]string{opt1, opt2}, 0, size, 0, "",
		"", "", "", f, nil))
}

func command(cmd string, size int, f arg.FuncExecutor) {
	wrap(arg.AddCommand([]string{cmd}, 0, size, "",
		"", "", "", f, nil))
}

func wrap(err error) {
	if err != nil {
		log.Println(err)
		os.Exit(0)
	}
}
