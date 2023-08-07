#!/bin/bash

# protoc --go_out=hub_protos_agents --go_opt=paths=source_relative \
#     --go-grpc_out=hub_protos_agents --go-grpc_opt=paths=source_relative \
#     protos/agents.proto


protoc \
  --proto_path=protos \
  --go_out=internal/proto/hub_agent_protos \
  --go_opt=paths=source_relative \
  --go_opt=Magents.proto=/internal/proto/hub_agent_protos \
  --go-grpc_opt=Magents.proto=/internal/proto/hub_agent_protos \
  --go-grpc_opt=paths=source_relative \
  --go-grpc_out=internal/proto/hub_agent_protos \
  protos/*.proto
