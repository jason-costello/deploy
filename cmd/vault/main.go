package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/vault/api"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s key1=value1 key2=value2 ...", os.Args[0])
	}

	// Vault connection details
	vaultAddress := "http://127.0.0.1:8200" // Replace with your Vault URL
	vaultToken := "root"                    // Replace with your Vault token

	// Create a client to interact with Vault
	config := &api.Config{
		Address: vaultAddress,
	}

	client, err := api.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create Vault client: %v", err)
	}

	// Set the Vault token
	client.SetToken(vaultToken)

	// Verify the client is authenticated
	_, err = client.Auth().Token().LookupSelf()
	if err != nil {
		log.Fatalf("Failed to authenticate to Vault: %v", err)
	}
	fmt.Println("Successfully authenticated to Vault.")

	// Define the path
	vaultPath := "secret/data/stormsync/development"
	// Parse key/value pairs from command line arguments
	data := make(map[string]interface{})
	for _, arg := range os.Args[1:] {
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			log.Fatalf("Invalid key/value pair: %s", arg)
		}
		data[kv[0]] = kv[1]
	}

	// Write the key/value pairs to Vault
	secretData := map[string]interface{}{
		"data": data,
	}

	_, err = client.Logical().Write(vaultPath, secretData)
	if err != nil {
		log.Fatalf("Failed to write data to %s: %v", vaultPath, err)
	}

	fmt.Printf("Data written to %s successfully!\n", vaultPath)

	for k := range data {
		val, err := readFromVault(client, vaultPath, k)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println("k: " + k + "  val: " + val)
	}

}
func readFromVault(client *api.Client, path, key string) (string, error) {
	secret, err := client.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read data from %s: %w", path, err)
	}

	if secret == nil || secret.Data["data"] == nil {
		return "", fmt.Errorf("no data found at %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data format error at %s", path)
	}

	value, ok := data[key].(string)
	if !ok {
		return "", fmt.Errorf("key %s not found at %s", key, path)
	}

	return value, nil
}
