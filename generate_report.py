import json

def parse_newman_report(report_path):
    with open(report_path, 'r') as f:
        report = json.load(f)
    
    entries = []
    if 'run' in report and 'executions' in report['run']:
        for execution in report['run']['executions']:
            if 'request' in execution:
                method = execution['request']['method']
                url = execution['request'].get('url', {})
                path = extract_path(url)
                
                status = "Not Covered"
                if 'response' in execution:
                    if 'status' in execution['response']:
                        status = execution['response']['status']
                    else:
                        status = "Failure"
                
                entries.append((method, path, status))
    
    return entries

def extract_path(url):
    # Extract path from the 'url' dictionary
    if url and 'path' in url:
        path_parts = url['path']
        if isinstance(path_parts, list) and path_parts:
            return '/' + '/'.join(path_parts)
    return ""

def print_report(entries):
    print("| # | METHOD | PATH | RESULT | SOURCE |")
    print("|---|--------|------|--------|--------|")
    for i, entry in enumerate(entries, 1):
        method, path, result = entry
        print(f"| {i} | {method} | {path} | {result} | ../newman-report.json |")

def main():
    report_path = "C:/Users/Rekanto/Desktop/employee-service/analyze/newman-report.json"
    entries = parse_newman_report(report_path)
    print_report(entries)

if __name__ == "__main__":
    main()
