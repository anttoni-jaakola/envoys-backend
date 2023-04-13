package stock

import (
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbstock"
)

// Service - The purpose of this code is to create a "Service" struct that contains a pointer to an assets.Context. This allows the
// service to access the context and any of the assets within the context.
type Service struct {
	Context *assets.Context
}

// queryAgent - This function is used to get an Agent based on the userId provided. It uses a SQL query to search for an Agent with
// the given userId and returns the Agent's details. It also handles errors in case there is no Agent with the given userId.
func (s *Service) queryAgent(userId int64) (*pbstock.Agent, error) {

	var (
		response pbstock.Agent
	)

	// This block of code is used to query a database and return information based on a userId as an input. The query looks
	// for a row in the "agents" table that matches the userId. If there is a match, the code will scan the row and store
	// the values in the "response" variable, which is then returned. If there is no match, an error is returned.
	if err := s.Context.Db.QueryRow("select a.id, a.user_id, case when a.broker_id > 0 then b.name else a.name end as agent_name, a.broker_id, a.type, a.status, a.create_at from agents a left join agents b on b.id = a.broker_id where a.user_id = $1", userId).Scan(&response.Id, &response.UserId, &response.Name, &response.BrokerId, &response.Type, &response.Status, &response.CreateAt); err != nil {
		return &response, err
	}

	return &response, nil
}
