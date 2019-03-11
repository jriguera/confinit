package config

import (
	//"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strconv"

	"confinit/pkg/fs"
	"confinit/pkg/log"

	validator "github.com/asaskevich/govalidator"
)

func init() {
	// validation to fail when struct fields do not include validations or are
	// not explicitly marked as exempt (using valid:"-" or
	// valid:"email,optional")
	validator.SetFieldsRequiredByDefault(false)
	validator.TagMap["mode"] = validator.Validator(validateMode)
	validator.TagMap["user"] = validator.Validator(validateUser)
	validator.TagMap["group"] = validator.Validator(validateGroup)
	validator.TagMap["glob"] = validator.Validator(validateGlob)
	// Structs
	validator.CustomTypeTagMap.Set("configuration",
		validator.CustomTypeValidator(
			func(i interface{}, context interface{}) bool {
				switch c := context.(type) {
				case Permissions:
					if err := c.Validate(); err == nil {
						return true
					}
				case Operation:
					if err := c.Validate(); err == nil {
						return true
					}
				}
				return false
			}))
}

// Validate Permissions
func (p *Permissions) Validate() error {
	_, err := validator.ValidateStruct(p)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func validateGlob(s string) bool {
	if _, err := fs.NewGlob(s); err != nil {
		log.Errorf("Invalid glob pattern '%s', %s", s, err.Error())
		return false
	}
	return true
}

func validateMode(s string) bool {
	// // Create a Temp File to check mode
	tmpFile, err := ioutil.TempFile(os.TempDir(), "fs-check-*")
	if err != nil {
		log.Errorf("Cannot create temporary file, %s", err.Error())
		return false
	}
	defer os.Remove(tmpFile.Name())
	mode, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		log.Errorf("Cannot parse mode, %s", err)
		return false
	}
	if err := tmpFile.Chmod(os.FileMode(mode)); err != nil {
		log.Errorf("Invalid mode, %s", err.Error())
		return false
	}
	return true
}

func validateUser(s string) bool {
	_, err := user.LookupId(s)
	if err != nil {
		_, err = user.Lookup(s)
		if err != nil {
			log.Errorf("Invalid %s", err.Error())
			return false
		}
	}
	return true
}

func validateGroup(s string) bool {
	_, err := user.LookupGroupId(s)
	if err != nil {
		_, err = user.LookupGroup(s)
		if err != nil {
			log.Errorf("Invalid %s", err.Error())
			return false
		}
	}
	return true
}

// Validate Operation
func (o *Operation) Validate() error {
	_, err := regexp.Compile(o.Regex)
	if err != nil {
		err = fmt.Errorf("Invalid pattern '%s', %s", o.Regex, err)
		log.Error(err)
		return err
	}
	if o.DestinationPath != "" {
		r, _ := validator.IsFilePath(o.DestinationPath)
		if !r {
			err := fmt.Errorf("Invalid path: %s", o.DestinationPath)
			log.Error(err)
			return err
		}
	}
	if o.Command != nil {
		_, err := validator.ValidateStruct(o.Command)
		if err != nil {
			log.Error(err)
			return err
		}
	}
	if o.DestinationPath == "" && o.Command == nil {
		return fmt.Errorf("Action not valid")
	}
	return nil
}

// Validate Config the application's configuration
func (c *Config) Validate() error {
	if _, err := validator.ValidateStruct(c); err != nil {
		log.Errorf("Configuration is not correct: %s", err)
		return err
	}
	log.Debug("Configuration format is correct")
	return nil
}
