import sys
import yaml

def generate_compose(num_clients: int):
    compose = {
        "name": "tp0",
        "services": {
            "server": {
                "container_name": "server",
                "image": "server:latest",
                "volumes": ["./server/config.ini:/config.ini"],

                "entrypoint": "python3 /main.py",
                "environment": [
                    "PYTHONUNBUFFERED=1",
                ],
                "networks": ["testing_net"],
            }
        },
        "networks": {
            "testing_net": {
                "ipam": {
                    "driver": "default",
                    "config": [
                        {"subnet": "172.25.125.0/24"}
                    ]
                }
            }
        }
    }

    for i in range(1, num_clients + 1):
        compose["services"][f"client{i}"] = {
            "container_name": f"client{i}",
            "image": "client:latest",
            "volumes": ["./client/config.yaml:/config.yaml", f"./data/dataset.csv:/data/agency-{i}.csv"],
            "entrypoint": "/client",
            "env_file": "./client/.env",
            "environment": [
                f"CLI_ID={i}",
            ],
            "networks": ["testing_net"],
            "depends_on": ["server"],
        }

    return compose


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Usage: python generate_compose.py <filename> <num_clients>")
        sys.exit(1)

    filename = sys.argv[1]
    try:
        num_clients = int(sys.argv[2])
    except ValueError:
        print("Error: num_clients must be an integer")
        sys.exit(1)

    data = generate_compose(num_clients)

    with open(filename, "w") as f:
        yaml.dump(data, f, sort_keys=False)

    print(f"{filename} generated with {num_clients} clients.")
