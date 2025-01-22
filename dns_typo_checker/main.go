package dns_typo_checker

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

// GenerateTypoDomains creates a list of typo variations for a domain
func GenerateTypoDomains(domain string, commonTLDs []string) []string {
	typos := []string{}
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 {
		return typos
	}

	name := domainParts[0]
	tld := strings.Join(domainParts[1:], ".")

	// Add common typos for the main domain name
	for i := 0; i < len(name); i++ {
		// Omit a character
		typos = append(typos, name[:i]+name[i+1:]+"."+tld)

		// Swap adjacent characters
		if i < len(name)-1 {
			swapped := name[:i] + string(name[i+1]) + string(name[i]) + name[i+2:]
			typos = append(typos, swapped+"."+tld)
		}
	}

	// Add common TLD typos
	for _, typoTLD := range commonTLDs {
		if typoTLD != tld {
			typos = append(typos, name+"."+typoTLD)
		}
	}

	return typos
}

// CheckDNS checks if a domain resolves to a valid DNS record
func CheckDNS(domain string) bool {
	_, err := net.LookupNS(domain)
	return err == nil
}

// GetDomainOwner uses the "whois" command to retrieve domain ownership information
func GetDomainOwner(domain string) string {
	cmd := exec.Command("whois", domain)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("Error retrieving WHOIS data for %s: %v", domain, err)
	}
	return string(output)
}

func Run(domains []string, commonTLDs []string) {

	if len(domains) == 0 {
		fmt.Println("No domains provided for typo check")
		return
	}

	if len(commonTLDs) == 0 {
		fmt.Println("Use default common TLDs")
		commonTLDs = []string{"com", "net", "org", "ne", "co", "cm", "om", "de"}
	}
	// Open log file
	logFile, err := os.Create("dns_typo_checker_details.log")
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer logFile.Close()
	// Open No DNS log file
	noDNSLogFile, err := os.Create("dns_typo_checker_not_registered.log")
	if err != nil {
		fmt.Println("Error creating log file:", err)
		return
	}
	defer noDNSLogFile.Close()

	fmt.Println("Searching for DNS typos...")
	logFile.WriteString("Starting DNS typo checks\n")

	for _, domain := range domains {
		fmt.Printf("\nChecking typos for domain: %s\n", domain)
		logFile.WriteString(fmt.Sprintf("\nChecking typos for domain: %s\n", domain))
		typos := GenerateTypoDomains(domain, commonTLDs)
		for _, typo := range typos {
			if CheckDNS(typo) {
				result := fmt.Sprintf("Valid DNS found for typo: %s\n", typo)
				fmt.Print(result)
				logFile.WriteString(result)
				ownerInfo := GetDomainOwner(typo)
				logFile.WriteString(fmt.Sprintf("Domain owner info for %s:\n%s\n", typo, ownerInfo))
			} else {
				result := fmt.Sprintf("No DNS record for: %s\n", typo)
				fmt.Print(result)
				logFile.WriteString(result)
				noDNSLogFile.WriteString(result)
			}
		}
	}

	fmt.Println("DNS typo check completed. Results written to dns_typo_checker.log")
	logFile.WriteString("DNS typo check completed.\n")
}
