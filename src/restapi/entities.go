package restapi

type DefaultHomepage struct {
	CorrelationID string `json:"id" bson:"_id" yaml:"id"`
	Timestamp     string `json:"timestamp" bson:"timestamp" yaml:"timestamp"`
	Message       string `json:"message" bson:"message" yaml:"message"`
}
