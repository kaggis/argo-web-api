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

package results

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"labix.org/v2/mgo/bson"

	"github.com/argoeu/argo-web-api/utils/authentication"
	"github.com/argoeu/argo-web-api/utils/config"
	"github.com/argoeu/argo-web-api/utils/mongo"
	"github.com/gorilla/mux"
)

// THIS CONTROLLER IS JUST A DEMO AND IS NOT SOMETHING THAT WORKS.
// TODO: WRITE AN ACTUAL CONTROLLER FOR AVAILABILITY

// ListEndpointGroupResults endpoint group availabilities according to the http request
func ListEndpointGroupResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {

	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	contentType := "text/xml"
	charset := "utf-8"
	//STANDARD DECLARATIONS END
	tenantDbConfig, err := authentication.AuthenticateTenant(r.Header, cfg)
	if err != nil {
		if err.Error() == "Unauthorized" {
			code = http.StatusUnauthorized
			return code, h, output, err
		}
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)

	input := endpointGroupResultQuery{
		Name:        vars["lgroup_name"],
		Granularity: urlValues.Get("granularity"),
		Format:      strings.ToLower(urlValues.Get("format")),
		StartTime:   urlValues.Get("start_time"),
		EndTime:     urlValues.Get("end_time"),
		Report:      vars["report_name"],
	}

	if r.Header.Get("format") == "json" {
		contentType = "application/json"
	}

	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := ReportInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	results := []EndpointGroupInterface{}

	ts, _ := time.Parse(zuluForm, input.StartTime)
	te, _ := time.Parse(zuluForm, input.EndTime)
	tsYMD, _ := strconv.Atoi(ts.Format(ymdForm))
	teYMD, _ := strconv.Atoi(te.Format(ymdForm))

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": tsYMD, "$lte": teYMD},
		"report": input.Report,
		"type":   vars["lgroup_type"],
	}

	if len(input.Name) > 0 {
		filter["name"] = input.Name
	}

	// Select the granularity of the search daily/monthly
	if len(input.Granularity) == 0 || strings.ToLower(input.Granularity) == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := DailyEndpointGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	} else if strings.ToLower(input.Granularity) == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := MonthlyEndpointGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	}
	// mongo.Find(session, tenantDbConfig.Db, "endpoint_group_ar", bson.M{}, "_id", &results)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	output, err = createEndpointGroupResultView(results, report, input.Format)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err
}

func prepareFilter(input endpointGroupResultQuery) bson.M {
	ts, _ := time.Parse(zuluForm, input.StartTime)
	te, _ := time.Parse(zuluForm, input.EndTime)
	tsYMD, _ := strconv.Atoi(ts.Format(ymdForm))
	teYMD, _ := strconv.Atoi(te.Format(ymdForm))

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": tsYMD, "$lte": teYMD},
		"report": input.Report,
	}

	if len(input.Name) > 0 {
		// filter["name"] = bson.M{"$in": input.Name}
		filter["name"] = input.Name
	}

	return filter
}

// DailyEndpointGroup query to aggregate daily results from mongodb
func DailyEndpointGroup(filter bson.M) []bson.M {
	// Mongo aggregation pipeline
	// Select all the records that match q
	// Project to select just the first 8 digits of the date YYYYMMDD
	// Sort by profile->supergroup->endpointGroup->datetime
	query := []bson.M{
		{"$match": filter},
		{"$project": bson.M{
			"date":         bson.M{"$substr": list{"$date", 0, 8}},
			"availability": 1,
			"reliability":  1,
			"unknown":      1,
			"uptime":       1,
			"downtime":     1,
			"type":         1,
			"report":       1,
			"supergroup":   1,
			"name":         1}},
		{"$sort": bson.D{
			{"report", 1},
			{"supergroup", 1},
			{"name", 1},
			{"type", 1},
			{"date", 1}}}}

	return query
}

// MonthlyEndpointGroup query to aggregate monthly results from mongodb
func MonthlyEndpointGroup(filter bson.M) []bson.M {

	// Mongo aggregation pipeline
	// Select all the records that match q
	// Group them by the first six digits of their date (YYYYMM), their supergroup, their endpointGroup, their profile, etc...
	// from that group find the average of the uptime, u, downtime
	// Project the result to a better format and do this computation
	// availability = (avgup/(1.00000001 - avgu))*100
	// reliability = (avgup/((1.00000001 - avgu)-avgd))*100
	// Sort the results by namespace->profile->supergroup->endpointGroup->datetime

	query := []bson.M{
		{"$match": filter},
		{"$group": bson.M{
			"_id": bson.M{
				"date":       bson.M{"$substr": list{"$date", 0, 6}},
				"name":       "$name",
				"supergroup": "$supergroup",
				"report":     "$report"},
			"type":      bson.M{"$fist": "$type"},
			"avguptime": bson.M{"$avg": "$uptime"},
			"avgunkown": bson.M{"$avg": "$unknown"},
			"avgdown":   bson.M{"$avg": "$downtime"}}},
		{"$project": bson.M{
			"date":       "$_id.date",
			"name":       "$_id.name",
			"report":     "$_id.report",
			"supergroup": "$_id.supergroup",
			"unknown":    "$avgunkown",
			"uptime":     "$avguptime",
			"downtime":   "$avgdown",
			"type":       1,
			"avguptime":  1,
			"avgunkown":  1,
			"avgdown":    1,
			"availability": bson.M{
				"$multiply": list{
					bson.M{"$divide": list{
						"$avguptime", bson.M{"$subtract": list{1.00000001, "$avgunkown"}}}},
					100}},
			"reliability": bson.M{
				"$multiply": list{
					bson.M{"$divide": list{
						"$avguptime", bson.M{"$subtract": list{bson.M{"$subtract": list{1.00000001, "$avgunkown"}}, "$avgdown"}}}},
					100}}}},
		{"$sort": bson.D{
			{"report", 1},
			{"supergroup", 1},
			{"type", 1},
			{"name", 1},
			{"date", 1}}}}

	return query
}
