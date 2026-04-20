#!/usr/bin/env python3
"""
Sync default resource types from resource-types-contrib repository.

This script:
1. Reads the sync configuration
2. Fetches resource type files from resource-types-contrib
3. Identifies files marked for default registration
4. Syncs those files to the local repository
5. Validates the synced files
"""

import os
import sys
import yaml
import requests
import base64
import json
from pathlib import Path
from typing import Dict, List, Set, Any


class ResourceTypeSync:
    def __init__(self, config_path: str):
        """Initialize the sync with configuration."""
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.github_token = os.environ.get('GITHUB_TOKEN')
        self.source_repo = os.environ.get('SOURCE_REPO', self.config['source']['repository'])
        self.dry_run = os.environ.get('DRY_RUN', 'false').lower() == 'true'
        
        self.target_dir = Path(self.config['target']['directory'])
        self.changes: List[str] = []
    
    def get_github_api_headers(self) -> Dict[str, str]:
        """Get headers for GitHub API requests."""
        headers = {
            'Accept': 'application/vnd.github.v3+json',
        }
        if self.github_token:
            headers['Authorization'] = f'token {self.github_token}'
        return headers
    
    def fetch_repository_tree(self) -> List[Dict[str, Any]]:
        """Fetch the repository tree from GitHub API."""
        branch = self.config['source']['branch']
        url = f"https://api.github.com/repos/{self.source_repo}/git/trees/{branch}?recursive=1"
        
        print(f"Fetching repository tree from {self.source_repo}...")
        response = requests.get(url, headers=self.get_github_api_headers())
        
        if response.status_code != 200:
            print(f"Error fetching repository tree: {response.status_code}")
            print(response.text)
            sys.exit(1)
        
        return response.json().get('tree', [])
    
    def fetch_file_content(self, file_path: str) -> str:
        """Fetch file content from GitHub."""
        branch = self.config['source']['branch']
        url = f"https://api.github.com/repos/{self.source_repo}/contents/{file_path}?ref={branch}"
        
        response = requests.get(url, headers=self.get_github_api_headers())
        
        if response.status_code != 200:
            print(f"Error fetching file {file_path}: {response.status_code}")
            return None
        
        content_data = response.json()
        content = base64.b64decode(content_data['content']).decode('utf-8')
        return content
    
    def matches_patterns(self, path: str, patterns: List[str]) -> bool:
        """Check if path matches any of the given patterns."""
        from fnmatch import fnmatch
        return any(fnmatch(path, pattern) for pattern in patterns)
    
    def should_sync_file(self, file_content: str) -> bool:
        """Determine if a file should be synced based on configuration."""
        strategy = self.config['sync']['strategy']
        
        if strategy == 'metadata':
            try:
                data = yaml.safe_load(file_content)
                metadata_field = self.config['sync']['metadataField']
                
                # Check if defaultRegistration is set to true
                return data.get(metadata_field, False) is True
            except yaml.YAMLError as e:
                print(f"Error parsing YAML: {e}")
                return False
        
        elif strategy == 'convention':
            # If using convention, all files in the specified path should be synced
            return True
        
        return False
    
    def validate_resource_type_manifest(self, content: str) -> bool:
        """Validate that the file is a proper resource type manifest."""
        if not self.config['validation']['enabled']:
            return True
        
        try:
            data = yaml.safe_load(content)
            required_fields = self.config['validation']['requiredFields']
            
            for field in required_fields:
                if field not in data:
                    print(f"Missing required field: {field}")
                    return False
            
            # Additional validation: ensure types is a dictionary
            if 'types' in data and not isinstance(data['types'], dict):
                print("Field 'types' must be a dictionary")
                return False
            
            return True
        except yaml.YAMLError as e:
            print(f"YAML validation error: {e}")
            return False
    
    def sync_files(self):
        """Main sync logic."""
        tree = self.fetch_repository_tree()
        base_path = self.config['source']['basePath']
        file_patterns = self.config['filePatterns']
        exclude_patterns = self.config['excludePatterns']
        
        # Create target directory if it doesn't exist
        if not self.dry_run:
            self.target_dir.mkdir(parents=True, exist_ok=True)
        
        synced_files: Set[str] = set()
        
        for item in tree:
            if item['type'] != 'blob':
                continue
            
            path = item['path']
            
            # Check if file is in the base path
            if not path.startswith(base_path):
                continue
            
            # Check if file matches include patterns
            if not self.matches_patterns(path, file_patterns):
                continue
            
            # Check if file matches exclude patterns
            if self.matches_patterns(path, exclude_patterns):
                continue
            
            # Fetch file content
            print(f"Checking file: {path}")
            content = self.fetch_file_content(path)
            
            if content is None:
                continue
            
            # Check if file should be synced
            if not self.should_sync_file(content):
                print(f"  Skipping (not marked for default registration)")
                continue
            
            # Validate the manifest
            if not self.validate_resource_type_manifest(content):
                print(f"  Warning: Validation failed for {path}, skipping")
                continue
            
            # Determine target file path
            relative_path = path[len(base_path):].lstrip('/')
            
            # Apply file prefix if configured
            file_prefix = self.config['target'].get('filePrefix', '')
            if file_prefix:
                # Split path into directory and filename
                path_parts = relative_path.rsplit('/', 1)
                if len(path_parts) == 2:
                    dir_part, file_part = path_parts
                    relative_path = f"{dir_part}/{file_prefix}{file_part}"
                else:
                    relative_path = f"{file_prefix}{relative_path}"
            
            target_file = self.target_dir / relative_path
            
            # Check if file has changed
            needs_update = True
            if target_file.exists():
                with open(target_file, 'r') as f:
                    existing_content = f.read()
                if existing_content == content:
                    needs_update = False
            
            if needs_update:
                if self.dry_run:
                    print(f"  Would sync to: {target_file}")
                    self.changes.append(f"- {relative_path}")
                else:
                    print(f"  Syncing to: {target_file}")
                    target_file.parent.mkdir(parents=True, exist_ok=True)
                    with open(target_file, 'w') as f:
                        f.write(content)
                    self.changes.append(f"- {relative_path}")
                
                synced_files.add(str(target_file.relative_to(Path.cwd())))
            else:
                print(f"  No changes needed")
        
        # Report summary
        print(f"\n{'[DRY RUN] ' if self.dry_run else ''}Sync completed")
        print(f"Files checked: {len([i for i in tree if i['type'] == 'blob'])}")
        print(f"Files synced: {len(synced_files)}")
        
        if self.changes:
            print("\nChanges:")
            for change in self.changes:
                print(change)
            
            # Set output for GitHub Actions
            changes_output = "\\n".join(self.changes)
            if os.environ.get('GITHUB_OUTPUT'):
                with open(os.environ['GITHUB_OUTPUT'], 'a') as f:
                    f.write(f"changes<<EOF\n")
                    f.write("\n".join(self.changes))
                    f.write("\nEOF\n")
        else:
            print("\nNo changes detected")
    
    def run(self):
        """Run the sync process."""
        try:
            print("Starting resource type sync...")
            print(f"Source: {self.source_repo}")
            print(f"Target: {self.target_dir}")
            print(f"Strategy: {self.config['sync']['strategy']}")
            print(f"Dry run: {self.dry_run}")
            print()
            
            self.sync_files()
            
            print("\n✅ Sync process completed successfully")
            return 0
            
        except Exception as e:
            print(f"\n❌ Sync process failed: {e}")
            import traceback
            traceback.print_exc()
            return 1


def main():
    config_file = os.environ.get('CONFIG_FILE', '.github/resource-type-sync-config.yaml')
    
    if not os.path.exists(config_file):
        print(f"Configuration file not found: {config_file}")
        sys.exit(1)
    
    syncer = ResourceTypeSync(config_file)
    sys.exit(syncer.run())


if __name__ == '__main__':
    main()
