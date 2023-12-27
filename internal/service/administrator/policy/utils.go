package policy

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func (s *adminPolicyService) validateRolesSubPolicies(subPoliciesJson *string) error {

	minSubPolicies := s.config.PolicyRules.RolesPoliciesMin
	maxSubPolicies := s.config.PolicyRules.RolesPoliciesMax

	if (*subPoliciesJson == "" || *subPoliciesJson == "{}") && minSubPolicies == 1 {
		return errors.New("at least one sub-policy is required")
	}

	var subPolicies []struct {
		Source    string `json:"source"`
		Action    string `json:"action"`
		Allowance string `json:"allowance"`
	}

	if err := json.Unmarshal([]byte(*subPoliciesJson), &subPolicies); err != nil {
		return err
	}

	if len(subPolicies) == 0 {
		return errors.New("at least one sub-policy is required")
	}

	unique := make(map[string]struct{})
	for _, sp := range subPolicies {
		key := fmt.Sprintf("%s-%s-%s", sp.Source, sp.Action, sp.Allowance)
		unique[key] = struct{}{}
	}

	var uniqueSubPolicies []struct {
		Source    string `json:"source"`
		Action    string `json:"action"`
		Allowance string `json:"allowance"`
	}
	for k := range unique {
		parts := strings.Split(k, "-")
		if len(parts) == 3 {
			uniqueSubPolicies = append(uniqueSubPolicies, struct {
				Source    string `json:"source"`
				Action    string `json:"action"`
				Allowance string `json:"allowance"`
			}{
				Source:    parts[0],
				Action:    parts[1],
				Allowance: parts[2],
			})
		}
	}

	if len(uniqueSubPolicies) > maxSubPolicies {
		return errors.New("exceeds maximum number of sub-policies." + fmt.Sprintf("Max sub-policies:%d ", maxSubPolicies))
	}

	uniqueJson, err := json.Marshal(uniqueSubPolicies)
	if err != nil {
		return err
	}

	*subPoliciesJson = string(uniqueJson)

	return nil
}
