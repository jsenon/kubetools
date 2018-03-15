package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type ConfigyamlClusters struct {
	NameClu string `yaml:"name"`
	Cluster struct {
		Server                   string `yaml:"server"`
		InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify"`
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
	} `yaml:"cluster"`
}

// type ConfigyamlContexts struct {
// 	NameUse string `yaml:"name"`
// 	User    struct {
// 		ClientCertificateData string `yaml:"client-certificate-data"`
// 		ClientKeyData         string `yaml:"naclient-key-datame"`
// 		UserName              string `yaml:"username"`
// 		Password              string `yaml:"password"`
// 	} `yaml:"user"`
// }
// type ConfigyamlUsers struct {
// 	NameCon string `yaml:"name"`
// 	Context struct {
// 		Cluster string `yaml:"cluster"`
// 		User    string `yaml:"user"`
// 	} `yaml:"context"`
// }

var outputFile string
var inputFile string

var inject = &cobra.Command{
	Use:   "inject",
	Short: "Inject k8s context to global config file",
	Long: `Inject kubernetes context in your global config file.
	               Complete documentation is available at https://github.com/jsenon/kubetool
	               After export Use kubectl config use-context YOURCONTEXT`,
	Run: func(cmd *cobra.Command, args []string) {

		res := &Config{}

		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}

		defaultfile := "/.kube/config"

		//Default Output to config
		outputFile = usr.HomeDir + defaultfile
		inputFile = usr.HomeDir + defaultfile

		file, err := os.Open(jsonfile)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close() // nolint: errcheck

		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		//debug

		err = json.Unmarshal([]byte(string(b)), &res)
		f, err := json.Marshal(res.Clusters[0])

		fmt.Println("\nMarshall JSON\n", string(f))

		result := ConfigyamlClusters{
			NameClu: res.Clusters[0].NameClu,
		}

		fmt.Println("\n Use new Struct:\n", result)
		result3, err := yaml.Marshal(result)
		fmt.Println("\n New struct in YAML\n", string(result3))

		myData3, err := yaml.Marshal(res.Clusters[0])
		fmt.Println("\nMarshall YAML:\n", string(myData3))

		myData4, err := yaml.Marshal(res)
		fmt.Println("\nMarshall ALL YAML:\n", string(myData4))

		out, _ := os.OpenFile("output.txt", os.O_APPEND|os.O_WRONLY, 0666)
		in, _ := os.Open(inputFile)

		fmt.Println("Write into:", outputFile)
		if err != nil {
			log.Fatal(err)
		}
		defer out.Close() // nolint: errcheck
		defer in.Close()

		scanner := bufio.NewScanner(in)

		writter := bufio.NewWriter(out)
		var line string

		for scanner.Scan() {
			line = scanner.Text()

			if idx := strings.Contains(line, "clusters:"); idx != false {

				fmt.Println("\nwhat we will write ton file:\n\n", string(myData3))

				// Check if we can delete field in live before write to buffer (clusters:)
				// Check if i can use temporary file already converted in yaml and use scanner bufio

				line = line + "\n" + string(myData3)
				writter.WriteString(line)
				writter.Flush()

			} else {
				if idx := strings.Contains(line, "contexts:"); idx != false {

					line = line + "\nINSERT CONT\n"
					writter.WriteString(line)
					writter.Flush()
				} else {
					if idx := strings.Contains(line, "users:"); idx != false {
						line = line + "\nINSERT USERS\n"
						writter.WriteString(line)
						writter.Flush()
					} else {
						line = line + "\n"
						writter.WriteString(line)
						writter.Flush()
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		//end debug

		// fmt.Println("***\n Import Your Context:", res.Clusters[0].Name)

		//Write to file asked by user
		// if cfgFile != "" {
		// 	e, err := yaml.Marshal(&res)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	err = ioutil.WriteFile(cfgFile, e, 0644)
		// } else {
		// 	// Write to default file
		// 	// f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY, 0600)
		// 	// if err != nil {
		// 	// 	log.Fatal(err)
		// }
		// defer f.Close()

		// d, err := yaml.Marshal(&res)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// fmt.Println("Data", string(d))
		// if _, err = f.WriteString(string(d)); err != nil {
		// 	panic(err)
		// }
		// }

	},
}

func init() {
	RootCmd.AddCommand(inject)

	inject.PersistentFlags().StringVarP(&jsonfile, "jsonimportfile", "i", "", "JSON input to add to your global context")
	inject.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "k8s config file output. Default ($HOME/.kube/config)")

	err := inject.MarkPersistentFlagRequired("jsonimportfile")
	if err != nil {
		log.Fatal(err)
	}

	err = viper.BindPFlag("jsonimportfile", extract.PersistentFlags().Lookup("jsonimportfile"))
	if err != nil {
		log.Fatal(err)
	}
	err = viper.BindPFlag("cfgFile", extract.PersistentFlags().Lookup("cfgFile"))
	if err != nil {
		log.Fatal(err)
	}
}
