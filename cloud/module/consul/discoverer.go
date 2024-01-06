package consul

type Datacenter string

const (
	SingleDatacenter Datacenter = "SINGLE"
	MultiDatacenter  Datacenter = "MULTI"
)

type DiscoveryClient struct {
	client *Client

	dc Datacenter
	// resolve service entry endpoints
	// resolver ServiceResolver
}

// Service get services from consul
// func (c *DiscoveryClient) Service(ctx context.Context, service string, index uint64, passingOnly bool) ([]*cloud.RemoteService, uint64, error) {
// 	if c.dc == MultiDatacenter {
// 		return c.multiDCService(ctx, service, index, passingOnly)
// 	}

// 	opts := &api.QueryOptions{
// 		WaitIndex:  index,
// 		WaitTime:   time.Second * 55,
// 		Datacenter: string(c.dc),
// 	}
// 	opts = opts.WithContext(ctx)

// 	if c.dc == SingleDatacenter {
// 		opts.Datacenter = ""
// 	}

// 	entries, meta, err := c.singleDCEntries(service, "", passingOnly, opts)
// 	if err != nil {
// 		return nil, 0, err
// 	}
// 	return c.resolver(ctx, entries), meta.LastIndex, nil
// }

// func (c *DiscoveryClient) multiDCService(ctx context.Context, service string, index uint64, passingOnly bool) ([]*cloud.RemoteService, uint64, error) {
// 	opts := &api.QueryOptions{
// 		WaitIndex: index,
// 		WaitTime:  time.Second * 55,
// 	}
// 	opts = opts.WithContext(ctx)

// 	var instances []*cloud.RemoteService

// 	dcs, err := c.client.Catalog().Datacenters()
// 	if err != nil {
// 		return nil, 0, err
// 	}

// 	for _, dc := range dcs {
// 		opts.Datacenter = dc
// 		e, m, err := c.singleDCEntries(service, "", passingOnly, opts)
// 		if err != nil {
// 			return nil, 0, err
// 		}

// 		ins := c.resolver(ctx, e)
// 		for _, in := range ins {
// 			if in.Metadata == nil {
// 				in.Metadata = make(map[string]string, 1)
// 			}
// 			in.Metadata["dc"] = dc
// 		}

// 		instances = append(instances, ins...)
// 		opts.WaitIndex = m.LastIndex
// 	}

// 	return instances, opts.WaitIndex, nil
// }

// func (c *DiscoveryClient) singleDCEntries(service, tag string, passingOnly bool, opts *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
// 	return c.cli.Health().Service(service, tag, passingOnly, opts)
// }

// // ServiceResolver is used to resolve service endpoints
// type ServiceResolver func(ctx context.Context, entries []*api.ServiceEntry) []*cloud.RemoteService

// func defaultResolver(_ context.Context, entries []*api.ServiceEntry) []*cloud.RemoteService {
// 	services := make([]*cloud.RemoteService, 0, len(entries))
// 	for _, entry := range entries {
// 		var version string
// 		for _, tag := range entry.Service.Tags {
// 			ss := strings.SplitN(tag, "=", 2)
// 			if len(ss) == 2 && ss[0] == "version" {
// 				version = ss[1]
// 			}
// 		}
// 		endpoints := make([]string, 0)
// 		for scheme, addr := range entry.Service.TaggedAddresses {
// 			if scheme == "lan_ipv4" || scheme == "wan_ipv4" || scheme == "lan_ipv6" || scheme == "wan_ipv6" {
// 				continue
// 			}
// 			endpoints = append(endpoints, addr.Address)
// 		}
// 		if len(endpoints) == 0 && entry.Service.Address != "" && entry.Service.Port != 0 {
// 			endpoints = append(endpoints, fmt.Sprintf("http://%s:%d", entry.Service.Address, entry.Service.Port))
// 		}
// 		services = append(services, &cloud.RemoteService{
// 			ID:       entry.Service.ID,
// 			Name:     entry.Service.Service,
// 			Metadata: entry.Service.Meta,
// 			Version:  version,
// 			// Endpoints: endpoints,
// 		})
// 	}

// 	return services
// }
