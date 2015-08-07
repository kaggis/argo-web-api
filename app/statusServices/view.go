/*
 * Copyright (c) 2014 GRNET S.A., SRCE, IN2P3 CNRS Computing Centre
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the
 * License. You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS
 * IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language
 * governing permissions and limitations under the License.
 *
 * The views and conclusions contained in the software and
 * documentation are those of the authors and should not be
 * interpreted as representing official policies, either expressed
 * or implied, of either GRNET S.A., SRCE or IN2P3 CNRS Computing
 * Centre
 *
 * The work represented by this source file is partially funded by
 * the EGI-InSPIRE project through the European Commission's 7th
 * Framework Programme (contract # INFSO-RI-261323)
 */

package statusServices

import (
	"encoding/xml"
)

func createView(results []DataOutput, input InputParams) ([]byte, error) {

	docRoot := &rootXML{}

	if len(results) == 0 {
		output, err := xml.MarshalIndent(docRoot, " ", "  ")
		return output, err
	}

	prevEndpointGroup := ""
	prevService := ""

	var ppEndpointGroup *endpointGroupXML
	var ppService *serviceXML

	for _, row := range results {

		if row.EndpointGroup != prevEndpointGroup && row.EndpointGroup != "" {
			endpointGroup := &endpointGroupXML{}
			endpointGroup.Name = row.EndpointGroup
			endpointGroup.GroupType = input.groupType
			docRoot.EndpointGroups = append(docRoot.EndpointGroups, endpointGroup)
			prevEndpointGroup = row.EndpointGroup
			ppEndpointGroup = endpointGroup
		}

		if row.Service != prevService && row.Service != "" {
			service := &serviceXML{}
			service.Name = row.Service
			service.GroupType = "service"
			ppEndpointGroup.Services = append(ppEndpointGroup.Services, service)

			prevService = row.Service
			ppService = service
		}

		status := &statusXML{}
		status.Timestamp = row.Timestamp
		status.Value = row.Status
		ppService.Statuses = append(ppService.Statuses, status)

	}

	output, err := xml.MarshalIndent(docRoot, " ", "  ")
	return output, err

}

func messageXML(answer string) ([]byte, error) {
	docRoot := &message{}
	docRoot.Message = answer
	output, err := xml.MarshalIndent(docRoot, " ", "  ")
	return output, err
}
