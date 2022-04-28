# FLUID_TEST

## Pre-requisites

* Install [Golang](https://go.dev/)
* Install [LevelDb](https://github.com/golang/leveldb)
* APIs Testing Tools [Postman](https://www.postman.com/)
* Apache HTTP Server Benchmarking Tool [ab](https://httpd.apache.org/docs/2.4/programs/ab.html)

## Build and Test

### FLUID Model

Add and Configure Nodes:

```shell
cd test
mkdir testn
touch boot.json
```

* boot.json

  ```json
  {
    "ip": "",
    "p2pPort": "",
    "httpPort": "",
    "bootNodes": [  // Active connection
      ""
    ],
    "ProducerAddress": [  // Miners' Address in Chain-based Model
      "1CEKaGFuue6RSvS25BEwfnm6PDvDa1WDzN",
      "1KA4hRhigbDNpLfaziy3u3hYX8feZNsmKs",
      "17UqkJVLCR83kUfVXgAVMmCfxpYq7wPzvT",
      "16Nxsd36KjXs6LctFemRPYdwdZbeGHjeVh",
      "1BqQoXNeZXiRZVVTeQLDijDR9XoayjrQaR"
    ]
  }
  ```

Start Nodes:

```
./run.sh
```

### EOS Model

```shell
cd eos_test
```

Add and Configure Nodes:

```shell
cd test
mkdir testn
touch boot.json
```

* boot.json

  ```json
  {
    "ip": "",
    "p2pPort": "",
    "httpPort": "",
    "bootNodes": [  // Active connection
      ""
    ],
    "ProducerAddress": [  // Miners' Address in Chain-based Model
      "1CEKaGFuue6RSvS25BEwfnm6PDvDa1WDzN",
      "1KA4hRhigbDNpLfaziy3u3hYX8feZNsmKs",
      "17UqkJVLCR83kUfVXgAVMmCfxpYq7wPzvT",
      "16Nxsd36KjXs6LctFemRPYdwdZbeGHjeVh",
      "1BqQoXNeZXiRZVVTeQLDijDR9XoayjrQaR"
    ]
  }
  ```

Start Nodes:

```shell
./run.sh
```

### Test Tool

* Codes for statistical measurement data

  ```
  cd test_tool
  # TPS
  singleChain.go
  dagExp.go
  
  #Latency
  latency.go
  proposeLatency.go
  ```

* Workload Submission Script

  ```
  ab_test.sh
  ```
