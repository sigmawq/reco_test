package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

func Extractor() {
	_period := config["extraction_period_seconds"]
	period, _ := strconv.ParseInt(_period, 10, 64)
	if period == 0 {
		log.Println("Period formatting was incorrect, using default=60")
		period = 60
	}

	log.Printf("Extractor period: %v seconds", period)
	for {
		users, err := AsanaGetUsers()
		if err != nil {
			log.Println("Failed to extract asana users:", err)
		} else {
			for _, user := range users {
				userJson, err := json.Marshal(user)
				if err != nil {
					log.Println(err)
					continue
				}

				name := fmt.Sprintf("users/%v.json", user.Gid)
				err = os.WriteFile(name, userJson, 0777)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}

		workspaces, err := AsanaGetWorkspaceList()
		if err != nil {
			log.Println("Failed to extract asana users:", err)
			continue
		}

		for _, workspace := range workspaces {
			projects, err := AsanaGetProjects(workspace.Gid)
			if err != nil {
				log.Println("Failed to query projects for workspace gid", workspace.Gid)
				continue
			}

			for _, project := range projects {
				projectJson, err := json.Marshal(project)
				if err != nil {
					log.Println(err)
					continue
				}

				name := fmt.Sprintf("projects/%v.json", project.Gid)
				err = os.WriteFile(name, projectJson, 0777)
				if err != nil {
					log.Println(err)
					continue
				}
			}
		}

		time.Sleep(time.Duration(period) * time.Second)
	}
}
