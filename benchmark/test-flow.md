
### Steps

1. Log into coinbase
2. Update repo (`git pull`)
3. Navigate to ./hotcertification/benchmark
4. Check setup script (!correct node!) and run: `./scripts/setup_testbed.sh`
5. SSH into the node
6. Navigate to hotcertification/benchmark directory
7. `./scripts/setup_hotcertification.sh <dir_for_keys> <num_nodes>` 
8. `./scripts/start_servers <keys_directory> <num_nodes>`
9. `./scripts/stop_servers`

### TODO

- [ ] add arguments to keygen.sh so it can scale to n nodes
  - [ ] write configuration file automatically for n nodes
  - [x] combine `keygen.sh` and `key-org.sh` into one script
  - [ ] add error msg if no args
  - [ ] calculate threshold t (e.g. n/3+1)

- [ ] write actual test cases and start client in a docker container
  - what am I measuring?
  - how am I measuring? directly in client go implementation? access hc from home computer?
- [ ] write down evaluation strategy