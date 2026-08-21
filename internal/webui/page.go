package webui

import "os"

const Page = `<!doctype html>
<html lang="zh-CN"><head><meta charset="utf-8"><title>TSN Slot Lab</title>
<style>body{font:16px system-ui;margin:2rem;background:#101820;color:#e9f1f7}button{padding:.6rem 1rem;margin:.3rem;background:#2b90d9;color:#fff;border:0;border-radius:4px}pre{background:#192934;padding:1rem;overflow:auto}code{color:#9fe870}</style></head>
<body><h1>TSN 时隙调度与抖动验证</h1><p>这是服务端调度能力的最小操作页面，不承担业务状态。</p>
<button onclick="load('/api/healthz')">健康</button><button onclick="load('/api/topology')">拓扑</button><button onclick="load('/api/drafts')">草稿</button><button onclick="load('/api/stats')">统计</button><pre id="out">请选择一个操作</pre>
<script>async function load(path){const r=await fetch(path);document.getElementById('out').textContent=JSON.stringify(await r.json(),null,2)}</script></body></html>`

func Load(path string) string {
	if b, err := os.ReadFile(path); err == nil && len(b) > 0 {
		return string(b)
	}
	return Page
}
