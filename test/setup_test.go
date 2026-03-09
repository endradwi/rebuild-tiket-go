package controllers_test

import (
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// init ensures the .env file from the project root is loaded before tests run
func init() {
    _, b, _, _ := runtime.Caller(0)
    basepath := filepath.Dir(b)
    
    // Move up one directory from controllers to root to find .env
    envPath := filepath.Join(basepath, "..", ".env")
    godotenv.Load(envPath)
}
