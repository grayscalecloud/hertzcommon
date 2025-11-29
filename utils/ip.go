// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"errors"
	"net"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

func GetLocalIPv4() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != net.FlagLoopback && iface.Flags&net.FlagUp != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
					return ipNet.IP.String(), nil
				}
			}
		}
	}
	return "", errors.New("get local IP error")
}

func MustGetLocalIPv4() string {
	ipv4, err := GetLocalIPv4()
	if err != nil {
		panic("get local IP error")
	}
	return ipv4
}

func AddLocalIpv4(addr string) string {
	if strings.HasPrefix(addr, ":") {
		addr = MustGetLocalIPv4() + addr
	}
	return addr
}

// GetClientIP 从 hertz 的 RequestContext 获取客户端 IP 地址
// 优先检查 X-Forwarded-For、X-Real-IP 等代理头，最后使用 RemoteAddr
func GetClientIP(c *app.RequestContext) string {
	if c == nil {
		return ""
	}

	// 优先使用 Hertz 内置的 ClientIP 方法，它会自动处理各种代理头
	ip := c.ClientIP()
	if ip != "" {
		return ip
	}

	// 如果 ClientIP 返回空，手动检查常见的代理头
	// 检查 X-Forwarded-For 头（可能包含多个 IP，取第一个）
	if xff := string(c.Request.Header.Get("X-Forwarded-For")); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip = strings.TrimSpace(ips[0])
			if ip != "" {
				return ip
			}
		}
	}

	// 检查 X-Real-IP 头
	if xri := string(c.Request.Header.Get("X-Real-IP")); xri != "" {
		ip = strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	// 检查 X-Forwarded 头
	if xf := string(c.Request.Header.Get("X-Forwarded")); xf != "" {
		ip = strings.TrimSpace(xf)
		if ip != "" {
			return ip
		}
	}

	// 最后使用 RemoteAddr
	remoteAddr := c.RemoteAddr().String()
	if remoteAddr != "" {
		// RemoteAddr 格式通常是 "IP:Port"，需要提取 IP 部分
		host, _, err := net.SplitHostPort(remoteAddr)
		if err == nil && host != "" {
			return host
		}
		// 如果没有端口（IPv6 或其他格式），直接返回
		return remoteAddr
	}

	return ""
}
