package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/panjf2000/ants/v2"
	"github.com/xuri/excelize/v2"
)

var (
	fileID     = kingpin.Flag("id", "文件相同属性").Short('i').Required().String()
	sourcePath = kingpin.Flag("source", "被分割文件地址(绝对地址)").Short('s').Required().String()
	destPath   = kingpin.Flag("dest", "分割文件存储地址,不带`/`是程序运行的相对地址").Short('d').Default("/home/ipcc/data/crm-import-tmp").String()
	goNum      = kingpin.Flag("num", "并发数").Short('c').Default("20").Int()
	goHandle   = kingpin.Flag("handle", "每个文件的数据量").Short('p').Default("10000").Int()
	debug      = kingpin.Flag("debug", "开启日志").Short('f').Bool()
)

func writeFile(row [][]string, n int, path *string) {
	fileName := strings.TrimRight(*path, "/") + "/" + strconv.Itoa(n) + ".json"
	dstFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println("open file ", n, " failed,err :", err)
	}

	st := time.Now()
	defer func() {
		log.Println("file ", n, " success ", time.Now().Sub(st).Seconds(), "s")
	}()

	jsonBytes, err := json.Marshal(row)
	if err != nil {
		log.Println("json file ", n, " failed,err :", err)
	}

	io.WriteString(dstFile, string(jsonBytes))
}

func main() {
	var (
		wg       sync.WaitGroup
		n        = 0
		rowIndex = 0
		st       = time.Now()
	)

	kingpin.Parse()

	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("fatal: %v", e)
		}
	}()

	path := strings.TrimRight(*destPath, "/") + "/" + *fileID
	if _, err := os.Stat(path); err != nil {
		if err = os.MkdirAll(path, 0711); err != nil {
			panic(fmt.Sprintf("Error creating directory: %v", err))
		}
	}

	if *debug == true {
		logFile, err := os.OpenFile(path+"/"+*fileID+".log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic(fmt.Sprintf("Error open log failed: %v", err))
		}
		log.SetOutput(logFile)
	}

	log.Println("start")

	f, err := excelize.OpenFile(*sourcePath)
	if err != nil {
		log.Panicln("Error open file failed", err)
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("Error close file failed", err)
		}
	}()

	sheetName := f.GetSheetName(f.GetActiveSheetIndex())
	rows, err := f.Rows(sheetName)
	if err != nil {
		log.Panicln("Error open file failed,err: ", err)
	}

	p, _ := ants.NewPoolWithFunc(*goNum, func(fileData interface{}) {
		mMap, ok := fileData.(map[string]any)
		if !ok {
			log.Println("Error interface assert failed")
			return
		}
		writeFile(mMap["row"].([][]string), mMap["num"].(int), &path)
		wg.Done()
	})
	defer ants.Release()
	defer p.Release()

	data := make([][]string, 0, *goHandle)

	for rows.Next() {
		if rowIndex > 0 {
			row, err := rows.Columns()
			if err != nil {
				log.Println("Error row failed, line:", rowIndex, " err:", err)
				continue
			}

			data = append(data, row)

			if rowIndex%*goHandle == 0 {
				tmpRow := make([][]string, *goHandle)
				copy(tmpRow, data)
				data = data[:0]
				wg.Add(1)
				_ = p.Invoke(map[string]any{"row": tmpRow, "num": n})
				n++
			}
		}
		rowIndex++
	}

	if len(data) > 0 {
		wg.Add(1)
		_ = p.Invoke(map[string]any{"row": data, "num": n})
	}

	if err = rows.Close(); err != nil {
		log.Println("Error rows close failed,", rowIndex, " err:", err)
	}

	wg.Wait()

	log.Println("success ", time.Now().Sub(st).Seconds(), "s")
}
