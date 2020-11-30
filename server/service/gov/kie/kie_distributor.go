package kie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-service-center/pkg/gov"
	"github.com/apache/servicecomb-service-center/server/config"
	svc "github.com/apache/servicecomb-service-center/server/service/gov"
	"github.com/ghodss/yaml"
	"github.com/go-chassis/kie-client"
	"log"
	"strings"
)

type Distributor struct {
	lbPolicies map[string]*gov.LoadBalancer
	name       string
	client     *kie.Client
}

const (
	PREFIX         = "servicecomb."
	EnableStatus   = "enabled"
	ValueType      = "text"
	AppKey         = "app"
	EnvironmentKey = "environment"
)

var rule = Validator{}

func (d *Distributor) Create(kind, project string, spec []byte) error {
	p := &gov.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("create %v", &p))
	key := toSnake(kind) + "." + p.Name
	err = rule.Validate(kind, p.Spec)
	if err != nil {
		return err
	}
	yamlByte, err := yaml.Marshal(p.Spec)
	if err != nil {
		return err
	}
	kv := kie.KVRequest{
		Key:       PREFIX + key,
		Value:     string(yamlByte),
		Status:    EnableStatus,
		ValueType: ValueType,
		Labels:    map[string]string{AppKey: p.Selector.App, EnvironmentKey: p.Selector.Environment},
	}
	_, err = d.client.Create(context.TODO(), kv, kie.WithProject(project))
	if err != nil {
		log.Fatal("kie create failed", err)
		return err
	}
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return nil
}

func (d *Distributor) Update(id, kind, project string, spec []byte) error {
	p := &gov.LoadBalancer{}
	err := json.Unmarshal(spec, p)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("update %v", &p))
	err = rule.Validate(kind, p.Spec)
	if err != nil {
		return err
	}
	yamlByte, err := yaml.Marshal(p.Spec)
	if err != nil {
		return err
	}
	kv := kie.KVRequest{
		ID:     id,
		Value:  string(yamlByte),
		Status: p.Status,
	}
	_, err = d.client.Put(context.TODO(), kv, kie.WithProject(project))
	if err != nil {
		log.Fatal("kie update failed", err)
		return err
	}
	d.lbPolicies[p.GovernancePolicy.Name] = p
	return nil
}

func (d *Distributor) Delete(id, project string) error {
	err := d.client.Delete(context.TODO(), id, kie.WithProject(project))
	if err != nil {
		log.Fatal("kie delete failed", err)
		return err
	}
	return nil
}

func (d *Distributor) List(kind, project, app, env string) ([]byte, error) {
	list, _, err := d.client.List(context.TODO(),
		kie.WithKey("beginWith("+PREFIX+toSnake(kind)+")"),
		kie.WithLabels(map[string]string{AppKey: app, EnvironmentKey: env}),
		kie.WithRevision(0),
		kie.WithGetProject(project))
	if err != nil {
		return nil, err
	}
	r := make([]*gov.LoadBalancer, 0, list.Total)
	for _, item := range list.Data {
		goc := &gov.LoadBalancer{
			GovernancePolicy: &gov.GovernancePolicy{},
		}
		spec := make(map[string]interface{})
		specJSON, _ := yaml.YAMLToJSON([]byte(item.Value))
		err = json.Unmarshal(specJSON, &spec)
		if err != nil {
			log.Fatal("kie list failed", err)
			return nil, err
		}
		goc.ID = item.ID
		goc.Status = item.Status
		goc.Name = item.Key
		goc.Spec = spec
		goc.Selector.App = item.Labels[AppKey]
		goc.Selector.Environment = item.Labels[EnvironmentKey]
		goc.CreatTime = item.CreatTime
		goc.UpdateTime = item.UpdateTime
		r = append(r, goc)
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return b, nil
}

func (d *Distributor) Get(id, project string) ([]byte, error) {
	kv, err := d.client.Get(context.TODO(), id, kie.WithGetProject(project))
	if err != nil {
		log.Fatal("kie get failed", err)
		return nil, err
	}
	goc := &gov.LoadBalancer{
		GovernancePolicy: &gov.GovernancePolicy{},
	}
	goc.ID = kv.ID
	goc.Status = kv.Status
	goc.Name = kv.Key
	goc.Spec = kv
	goc.Selector.App = kv.Labels[AppKey]
	goc.Selector.Environment = kv.Labels[EnvironmentKey]
	goc.CreatTime = kv.CreatTime
	goc.UpdateTime = kv.UpdateTime
	b, _ := json.MarshalIndent(goc, "", "  ")
	return b, nil
}

func (d *Distributor) Type() string {
	return svc.ConfigDistributorKie
}
func (d *Distributor) Name() string {
	return d.name
}

func initClient(endpoint string) *kie.Client {
	client, err := kie.NewClient(
		kie.Config{Endpoint: endpoint,
			DefaultLabels: map[string]string{},
		})
	if err != nil {
		log.Fatalf("init kie client failed, err: %s", err)
	}
	return client
}

func new(opts config.DistributorOptions) (svc.ConfigDistributor, error) {
	return &Distributor{name: opts.Name, lbPolicies: map[string]*gov.LoadBalancer{}, client: initClient(opts.Endpoint)}, nil
}

func toSnake(name string) string {
	if name == "" {
		return ""
	}
	temp := strings.Split(name, "-")
	var buffer bytes.Buffer
	for num, v := range temp {
		vv := []rune(v)
		if num == 0 {
			buffer.WriteString(string(vv))
			continue
		}
		if len(vv) > 0 {
			if vv[0] >= 'a' && vv[0] <= 'z' { //首字母大写
				vv[0] -= 32
			}
			buffer.WriteString(string(vv))
		}
	}
	return buffer.String()
}

func init() {
	svc.InstallDistributor(svc.ConfigDistributorKie, new)
}
