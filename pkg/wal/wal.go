package wal

import (
	"bufio"
	"fmt"
	"os"

	t "github.com/mana-sg/kv-log-store/types"
	"github.com/mana-sg/kv-log-store/utils"
)

var LOGFILE string = "/kls/log.bin"

func WriteLog(op, key, value string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %v", err)
	}

	PATH := homeDir + LOGFILE

	log := CreateLog(op, key, value)
	encodedLog, err := utils.EncodeLog(log)
	if err != nil {
		return fmt.Errorf("error encoding log in WriteLog: %v", err)
	}

	file, err := os.OpenFile(PATH, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error reading log file in WriteLog: %v", err)
	}
	defer file.Close()

	_, err = file.Write(encodedLog)
	if err != nil {
		return fmt.Errorf("error writing to log file in WriteLog: %v", err)
	}

	_, err = file.Write([]byte("\n"))
	if err != nil {
		return fmt.Errorf("error writing newline to log file in WriteLog: %v", err)
	}

	return nil
}

func GetLogs() ([]t.LogEntry, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home directory: %v", err)
	}

	PATH := homeDir + LOGFILE

	var logs []t.LogEntry

	if _, err := os.Stat(PATH); os.IsNotExist(err) {
		f, err := os.Create(PATH)
		defer f.Close()
		if err != nil {
			return nil, fmt.Errorf("error creating log file in GetLogs: %v", err)
		}
		return logs, nil
	}

	file, err := os.Open(PATH)
	if err != nil {
		return nil, fmt.Errorf("error reading log file in GetLogs: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		log, err := utils.DecodeLog([]byte(line))
		if err != nil {
			return nil, fmt.Errorf("error decoding log in GetLogs: %v", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func Compact() (float64, error) {
	compactionHelper := make(map[string]int)
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, fmt.Errorf("unable to get user home directory: %v", err)
	}
	
	PATH := home + LOGFILE
	fileInfo, err := os.Stat(PATH)
	if err != nil {
		return 0, fmt.Errorf("unable to stat file: %v", err)
	}

	beforeSize := fileInfo.Size()
	file, err := os.Open(PATH)
	if err != nil {
		return 0, fmt.Errorf("error reading log file in Compact: %v", err)
	}
	defer file.Close()

	var logs []string
	scanner := bufio.NewScanner(file)
	lineNo := 0

	for scanner.Scan() {
		line := scanner.Text()
		log, err := utils.DecodeLog([]byte(line))
		if err != nil {
			continue 
		}
		logs = append(logs, line)
		compactionHelper[log.Key] = lineNo
		lineNo++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error scanning log file: %v", err)
	}

	tempFilePath := PATH + ".tmp"
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return 0, fmt.Errorf("unable to create temp log file: %v", err)
	}
	defer tempFile.Close()

	for _, idx := range compactionHelper {
		_, err := tempFile.WriteString(logs[idx] + "\n")
		if err != nil {
			return 0, fmt.Errorf("error writing compacted log: %v", err)
		}
	}

	err = os.Rename(tempFilePath, PATH)
	if err != nil {
		return 0, fmt.Errorf("failed to replace log file: %v", err)
	}

	afterInfo, err := os.Stat(PATH)
	if err != nil {
		return 0, fmt.Errorf("unable to stat compacted file: %v", err)
	}
	afterSize := afterInfo.Size()

	savings := float64(beforeSize-afterSize) / float64(beforeSize)
	return savings, nil
}

func CreateLog(op, key, value string) t.LogEntry {
	return t.LogEntry{
		Operation: op,
		Key:       key,
		Value:     value,
	}
}
