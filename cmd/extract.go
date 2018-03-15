package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config Struct of JSON Kubernetes Config File
type Config struct {
	Kind        string `json:"kind" yaml:"-"`
	APIVersion  string `json:"apiVersion" yaml:"-"`
	Preferences struct {
	} `json:"preferences" yaml:"-"`
	Clusters []struct {
		NameClu string `json:"name" yaml:"name"`
		Cluster struct {
			Server                   string `json:"server" yaml:"server"`
			InsecureSkipTLSVerify    bool   `json:"insecure-skip-tls-verify,omitempty" yaml:"insecure-skip-tls-verify"`
			CertificateAuthorityData string `json:"certificate-authority-data" yaml:"certificate-authority-data"`
		} `json:"cluster" yaml:"cluster"`
	} `json:"clusters" yaml:"clusters"`
	Users []struct {
		NameUs string `json:"name" yaml:"name"`
		User   struct {
			ClientCertificateData string `json:"client-certificate-data" yaml:"client-certificate-data"`
			ClientKeyData         string `json:"client-key-data" yaml:"naclient-key-datame"`
			UserName              string `json:"username" yaml:"username"`
			Password              string `json:"password" yaml:"password"`
		} `json:"user" yaml:"user"`
		AsUserExtra struct {
		} `json:"as-user-extra" yaml:"as-user-extra"`
	} `json:"users" yaml:"users"`
	Contexts []struct {
		NameCon string `json:"name" yaml:"name"`
		Context struct {
			Cluster string `json:"cluster" yaml:"cluster"`
			User    string `json:"user" yaml:"user"`
		} `json:"context" yaml:"context"`
	} `json:"contexts" yaml:"contexts"`
	CurrentContext string `json:"current-context" yaml:"-"`
}

var jsonfile string
var context string
var cfgFile string
var output string

var extract = &cobra.Command{
	Use:   "extract",
	Short: "Extract k8s context from global config file",
	Long: `Extract kubernetes context ie. configuration user and endpoint.
				  Complete documentation is available at https://github.com/jsenon/kubetools
				  After export Use kubectl config use-context YOURCONTEXT --kubeconfig output.json to use it`,
	// Args: cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {

		defaultfilejson := "/.kube/config.json"
		defaultfile := "/.kube/config"

		tempfile := ".convert.json"

		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		cmdName := "kubectl"
		cmdArgs := []string{"config", "view", "-o", "json", "--raw", "--kubeconfig", usr.HomeDir + defaultfile}

		// If no value for config k8s file, use default config but we need to convert to json
		if cfgFile == "" && jsonfile == "" {

			cfgFile = usr.HomeDir + defaultfile
			out, erro := exec.Command(cmdName, cmdArgs...).Output() // #nosec
			if erro != nil {
				log.Fatal(erro)
			}

			err = ioutil.WriteFile(tempfile, out, 0644)
			if err != nil {
				log.Fatal(err)
			}

			jsonfile = tempfile
		}

		// If value for config k8s but no json we need to generate a json output
		if cfgFile != "" && jsonfile == "" {

			cmdArgs := []string{"config", "view", "-o", "json", "--raw", "--kubeconfig", cfgFile}

			cfgFile = usr.HomeDir + defaultfile
			out, erro := exec.Command(cmdName, cmdArgs...).Output() // #nosecx
			if erro != nil {
				log.Fatal(erro)
			}

			err = ioutil.WriteFile(tempfile, out, 0644)
			if err != nil {
				log.Fatal(err)
			}

			jsonfile = tempfile

		}

		// Used default value for json config file
		// Exit if doesn't exist
		if jsonfile == "" {
			jsonfile = usr.HomeDir + defaultfilejson

		}

		file, err := os.Open(jsonfile)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close() // nolint: errcheck

		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		res := &Config{}
		var configoutput Config

		err = json.Unmarshal([]byte(string(b)), &res)
		if err != nil {
			log.Fatal(err)
		}

		configoutput.APIVersion = res.APIVersion
		configoutput.Kind = res.Kind

		// Loop over Clusters matching with context asked
		for _, coutput := range res.Clusters {

			if coutput.NameClu == context {
				configoutput.Clusters = append(configoutput.Clusters, coutput)

			}

		}

		// Loop over Users matching with contexy asked
		for _, coutput := range res.Users {

			if coutput.NameUs == context {
				configoutput.Users = append(configoutput.Users, coutput)
			}

		}

		// Loop over Contexts matching with contexy asked
		for _, coutput := range res.Contexts {

			if coutput.NameCon == context {
				configoutput.Contexts = append(configoutput.Contexts, coutput)
			}

		}

		// Output to console
		if output == "" {
			body, erro := json.MarshalIndent(configoutput, "", "   ")
			if erro != nil {
				log.Fatal(erro)
			}
			fmt.Println(string(body))
		} else {

			//Write to output file specified in args
			body, erro := json.MarshalIndent(configoutput, "", "   ")
			if erro != nil {
				log.Fatal(erro)
			}

			err = ioutil.WriteFile(output, body, 0644)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Exported to:", output)

		}

		// Delete temporary file
		err = os.Remove(tempfile)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Func to init Cobra Flag and bind flag
func init() {
	RootCmd.AddCommand(extract)

	extract.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "k8s config file default ($HOME/.kube/config)")
	extract.PersistentFlags().StringVarP(&jsonfile, "configjson", "j", "", "k8s config file JSON default ($HOME/.kube/config.json)")

	extract.PersistentFlags().StringVarP(&context, "context", "e", "", "MANDATORY: Name of  context to extract")
	extract.PersistentFlags().StringVarP(&output, "output", "o", "", "Name of output file")

	err := extract.MarkPersistentFlagRequired("context")
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("jsonfile", extract.PersistentFlags().Lookup("jsonfile"))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindPFlag("context", extract.PersistentFlags().Lookup("context"))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindPFlag("output", extract.PersistentFlags().Lookup("output"))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindPFlag("cfgFile", extract.PersistentFlags().Lookup("cfgFile"))
	if err != nil {
		log.Fatal(err)
	}
}
