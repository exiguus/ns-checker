package network

import (
	"fmt"
	"strconv"
)

const DefaultPort = 53

func ParsePort(port string) (int, error) {
	if port == "" {
		return DefaultPort, nil
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return 0, fmt.Errorf("port number must be between 1 and 65535")
	}

	return portNum, nil
}
