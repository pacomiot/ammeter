package cloudserver

//4G 网关
type FourGGatewayModel struct {
	AccessToken string `json:"accessToken"`
	MqttServer  string `json:"mqttserver"`
	Id          string `json:"id"`
	Name        string `json:"name"`
}
