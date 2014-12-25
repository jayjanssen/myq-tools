package main

import (
    "fmt"
    "./loader"
    // "./metricdefs"
    "reflect"
)

func main() {
  // Parse arguments
  
  // Load data
  samples, err := loader.GetSamplesFile("./loader/mysqladmin.lots")
  if err != nil {
    panic( err )
  }
  
  // Apply selected view to each sample
  first := <- samples
  fmt.Println( reflect.TypeOf( first["connections"] ))
  
}
