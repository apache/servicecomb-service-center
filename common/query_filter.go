//Copyright 2017 Huawei Technologies Co., Ltd
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
package common

import (
	"errors"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"net/url"
	"strconv"
	"strings"
)

type QueryFilter struct {
	Filter                 string
	Opt                    string
	Values                 []string
	Page                   int
	Results_per_page       int
	Order_asc              bool
	Inline_relations_depth int
}

const DEFAULT_PAGE = 1
const DEFAULT_RESULTS_PER_PAGE = 20

func ParseFilterPageInfo(queryForm url.Values, queryFilter *QueryFilter) error {
	var err error

	if len(queryForm["page"]) == 1 {
		queryFilter.Page, err = strconv.Atoi(queryForm["page"][0])
		if err != nil {
			return err
		} else if queryFilter.Page <= 0 {
			return errors.New("param page " + queryForm["page"][0] + " invalid.")
		}
	} else {
		queryFilter.Page = DEFAULT_PAGE
	}

	if len(queryForm["results-per-page"]) == 1 {
		queryFilter.Results_per_page, err = strconv.Atoi(queryForm["results-per-page"][0])
		if err != nil {
			return err
		} else if queryFilter.Results_per_page <= 0 {
			return errors.New("param results-per-page " + queryForm["results-per-page"][0] + " invalid.")
		}
	} else {
		queryFilter.Results_per_page = DEFAULT_RESULTS_PER_PAGE
	}

	return nil
}

func ParseFilterOrderInfo(queryForm url.Values, queryFilter *QueryFilter) error {
	if len(queryForm["order-direction"]) == 1 {
		if strings.EqualFold(queryForm["order-direction"][0], "desc") {
			queryFilter.Order_asc = false
		} else if strings.EqualFold(queryForm["order-direction"][0], "asc") {
			queryFilter.Order_asc = true
		} else {
			return errors.New("input param order_direction invalid, " + queryForm["order-direction"][0])
		}
	} else if len(queryForm["order-direction"]) == 0 {
		queryFilter.Order_asc = true
	} else {
		return errors.New("The order-direction param num is illegal")
	}
	return nil
}

func ParseCommQueryInfo(query string, queryFilter *QueryFilter) error {

	//delete prefix space
	query = strings.TrimPrefix(query, " ")
	if strings.HasPrefix(query, ":") {
		query = strings.TrimPrefix(query, ":")
		queryFilter.Opt = ""
	} else if strings.HasPrefix(query, ">=") {
		query = strings.TrimPrefix(query, ">=")
		queryFilter.Opt = "__gte"
	} else if strings.HasPrefix(query, "<=") {
		query = strings.TrimPrefix(query, "<=")
		queryFilter.Opt = "__lte"
	} else if strings.HasPrefix(query, "<") {
		query = strings.TrimPrefix(query, "<")
		queryFilter.Opt = "__lt"
	} else if strings.HasPrefix(query, ">") {
		query = strings.TrimPrefix(query, ">")
		queryFilter.Opt = "__gt"
	} else if strings.HasPrefix(query, "IN") {
		query = strings.TrimPrefix(query, "IN")
		queryFilter.Opt = "__in"
	} else {
		//beego.Error("Get all services failed:", "input param query opt invalid, ", query)
		return errors.New("input param query filter invalid, " + query)
	}

	//delete all space
	query = strings.Trim(query, " ")
	//split by ","
	queryFilter.Values = strings.Split(query, ",")
	if len(queryFilter.Values) == 0 ||
		(queryFilter.Opt != "__in" && len(queryFilter.Values) > 1) {

		//beego.Error("Get all services failed:", "input param  query value invalid, ", query)
		return errors.New("input param query filter invalid, " + query)
	}
	return nil
}

func GetTotalPage(record_cnt int, queryFilter *QueryFilter) (total_page_num int) {

	//default total page num is 1

	if record_cnt%queryFilter.Results_per_page == 0 {
		total_page_num = record_cnt / queryFilter.Results_per_page
	} else {
		total_page_num = record_cnt/queryFilter.Results_per_page + 1
	}

	if 0 == record_cnt {
		total_page_num = 1
	}

	//fmt.Println("total_page_num:",total_page_num  ," record_cnt :",record_cnt,"Results_per_page:",queryFilter.Results_per_page)
	return total_page_num
}

func QueryFilterAndOrder(qs orm.QuerySeter, queryFilter *QueryFilter) orm.QuerySeter {

	//set no limit record, beego return 1000 record by default
	qs.Limit(-1)
	if queryFilter.Filter != "" {
		filter := queryFilter.Filter + queryFilter.Opt
		qs = qs.Filter(filter, queryFilter.Values)

		var order string
		if queryFilter.Order_asc == true {
			order = queryFilter.Filter
		} else {
			order = "-" + queryFilter.Filter
		}
		qs = qs.OrderBy(order)
		beego.Debug("get service query filter:", filter, " values:",
			queryFilter.Values, " order:", order)
	}

	return qs
}
