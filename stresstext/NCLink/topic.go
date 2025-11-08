package NCLink

import "fmt"

func QueryRequestTopic(deviceID string) string {
	return fmt.Sprintf("Query/Request/%s", deviceID);
}

func QueryResponseTopic(deviceID string) string {
	return fmt.Sprintf("Query/Response/%s", deviceID);
}

func SetRequestTopic(deviceID string) string {
	return fmt.Sprintf("Set/Request/%s", deviceID);
}

func SetResponseTopic(deviceID string) string {
	return fmt.Sprintf("Set/Response/%s", deviceID);
}

func ProbeQueryRequestTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Query/Request/%s", deviceID);
}

func ProbeQueryResponseTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Query/Response/%s", deviceID);
}

func ProbeSetRequestTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Set/Request/%s", deviceID);
}

func ProbeSetResponseTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Set/Response/%s", deviceID);
}

func RegisterRequestTopic(deviceID string) string {
	return fmt.Sprintf("Register/Request");
}

func RegisterResponseTopic(deviceID string) string {
	return fmt.Sprintf("Register/Response/%s", deviceID);
}
func ProbeVersionTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Version/%s", deviceID)
}
func ProbeVersionResponseTopic(deviceID string) string {
	return fmt.Sprintf("Probe/Version/Response/%s", deviceID)
}
