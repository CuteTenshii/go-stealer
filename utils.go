package main

func humanizeBoolean(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}
