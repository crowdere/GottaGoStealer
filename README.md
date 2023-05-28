# GottaGoStealer
Blacktail/Buhti Ransomware Custom Exfiltration tool

I recreated the stealer described within a recent blog post by Symantec Threat Intelligence Team found here https://symantec-enterprise-blogs.security.com/blogs/threat-intelligence/buhti-ransomware.

This was never suppose to be good programing practice, nor extensivley created. I just wanted to try recreate and extend that project. 

See full blog post write up here: MEDIUM LINK


Tested on Windows 10 and MacOS Venture 13.4. It should work on Linux.

![alt text](images/6.Favicon%20stealer.png)

## how to build
`env GOOS=windows GOARCH=amd64 go build app.go`

## Command Line Arguments
`-o is the output file flag. output.zip should be used for the SaaS backend to work correctly.`

`-d  The search directory to begin crawling. This will crawl all sub folders from here `

`-e the SaaS backend endpoint public address for file exfiltration`

`-c the client ID to identify which client the data was sent from.`