package data

import (
	"Parallels/pd-api-service/data/models"
	"encoding/json"
	"errors"
	"os"
)

type Data struct {
	Users                   []models.User                   `json:"users"`
	Claims                  []models.UserClaim              `json:"claims"`
	Roles                   []models.UserRole               `json:"roles"`
	ApiKeys                 []models.ApiKey                 `json:"api_keys"`
	VirtualMachineTemplates []models.VirtualMachineTemplate `json:"virtual_machine_templates"`
}

type JsonDatabase struct {
	connected bool
	filename  string
	data      Data
}

func NewJsonDatabase(filename string) *JsonDatabase {
	return &JsonDatabase{
		connected: false,
		filename:  filename,
		data:      Data{},
	}
}

func (j *JsonDatabase) Connect() error {
	var data Data

	// Check if file exists, create it if it doesn't
	if _, err := os.Stat(j.filename); os.IsNotExist(err) {
		file, err := os.Create(j.filename)
		if err != nil {
			return err
		}
		file.Close()
	}

	file, err := os.Open(j.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := os.Stat(j.filename)
	if err != nil {
		return err
	}

	if fileInfo.Size() == 0 {
		j.data = Data{
			Users: make([]models.User, 0),
		}

		err = j.save()
		if err != nil {
			return err
		}
		j.connected = true
		return nil
	} else {
		// file is not empty

		decoder := json.NewDecoder(file)
		err = decoder.Decode(&data)
		if err != nil {
			return err
		}

		j.data = data
		j.connected = true
		return nil
	}
}

func (j *JsonDatabase) Disconnect() error {
	j.save()
	j.connected = false

	return nil
}

func (j *JsonDatabase) Filename() string {
	return j.filename
}

func (j *JsonDatabase) IsConnected() bool {
	return j.connected
}

func (j *JsonDatabase) save() error {
	if j.filename == "" {
		return errors.New("the database filename is not set")
	}

	file, err := os.OpenFile(j.filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonString, err := json.MarshalIndent(j.data, "", "  ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonString)
	if err != nil {
		return err
	}

	return nil
}
