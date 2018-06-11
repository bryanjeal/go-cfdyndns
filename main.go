package main

import (
	"errors"
	"net"

	cloudflare "github.com/cloudflare/cloudflare-go"
	"github.com/golang/glog"
)

const app = "go-cfdyndns"

func main() {
	// configure app
	appConf := configureApp(nil)

	// Construct a new API object
	api, err := cloudflare.New(appConf.GetString("cloudflare.apiKey"), appConf.GetString("cloudflare.apiEmail"))
	if err != nil {
		glog.Fatal(err)
	}

	// get zone ID
	zoneID, err := api.ZoneIDByName(appConf.GetString("app.zoneName"))
	if err != nil {
		glog.Fatal(err)
	}

	// base record to filter by
	record := cloudflare.DNSRecord{
		Name:     appConf.GetString("app.recordName"),
		Proxied:  false,
		Type:     "A",
		ZoneID:   zoneID,
		ZoneName: appConf.GetString("app.zoneName"),
	}

	// see if record exists
	recs, err := api.DNSRecords(zoneID, record)
	if err != nil {
		glog.Fatal(err)
	}

	// do we need to create a record or did we find one?
	if len(recs) > 0 {
		record.ID = recs[0].ID
	}

	// get IP
	ip, err := getIP()
	if err != nil {
		glog.Fatal(err)
	}

	// update cloudflare
	oldIP := record.Content
	if oldIP != ip {
		record.Content = ip
		if len(record.ID) > 0 {
			// update existing
			err := api.UpdateDNSRecord(zoneID, record.ID, record)
			if err != nil {
				glog.Fatal(err)
			} else {
				glog.Info("IP Updated to " + ip + ". Was " + oldIP)
			}
		} else {
			// create new
			resp, err := api.CreateDNSRecord(zoneID, record)
			if err != nil {
				glog.Fatal(err)
			} else if resp.Success {
				glog.Info("A Record Created. IP Set to " + ip)
			} else {
				glog.Fatal(resp.Errors)
			}
		}
	}
}

func getIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}
