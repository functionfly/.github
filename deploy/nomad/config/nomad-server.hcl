data_dir  = "/opt/nomad/data"
bind_addr = "0.0.0.0"

log_level = "INFO"

server {
  enabled          = true
  bootstrap_expect = 3
  server_affinity  = true
  num_schedulers   = 4
}

client {
  enabled = false
}

advertise {
  http = "{{ GetInterfaceIP \"eth0\" }}:4646"
  rpc  = "{{ GetInterfaceIP \"eth0\" }}:4647"
  serf = "{{ GetInterfaceIP \"eth0\" }}:4648"
}

enable_debug = false

skip_leave_on_interrupt = true
leave_on_interrupt = false

consul {
  address = "127.0.0.1:8500"
}