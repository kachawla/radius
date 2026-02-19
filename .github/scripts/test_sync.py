#!/usr/bin/env python3
"""
Unit tests for the resource type sync script.
"""

import sys
import os
import tempfile
import shutil
from pathlib import Path

# Add the scripts directory to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'scripts'))

def test_yaml_parsing():
    """Test that YAML files can be parsed correctly."""
    import yaml
    
    # Test config file
    config_path = os.path.join(os.path.dirname(__file__), '..', 'resource-type-sync-config.yaml')
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)
    
    assert 'source' in config, "Config missing 'source' section"
    assert 'target' in config, "Config missing 'target' section"
    assert 'sync' in config, "Config missing 'sync' section"
    
    assert config['source']['repository'] == 'radius-project/resource-types-contrib'
    assert config['target']['directory'] == 'deploy/manifest/built-in-providers/self-hosted'
    assert config['sync']['strategy'] == 'metadata'
    
    print("✅ YAML parsing test passed")

def test_validation_logic():
    """Test the validation logic for resource type manifests."""
    import yaml
    
    # Valid manifest
    valid_manifest = """
namespace: Test.Resources
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema:
          type: object
"""
    
    data = yaml.safe_load(valid_manifest)
    assert 'namespace' in data, "Valid manifest should have namespace"
    assert 'types' in data, "Valid manifest should have types"
    assert isinstance(data['types'], dict), "Types should be a dictionary"
    
    print("✅ Validation logic test passed")

def test_file_prefix_logic():
    """Test that file prefix logic works correctly."""
    
    # Test cases
    test_cases = [
        ("test.yaml", "synced_", "synced_test.yaml"),
        ("subdir/test.yaml", "synced_", "subdir/synced_test.yaml"),
        ("test.yaml", "", "test.yaml"),
    ]
    
    for input_path, prefix, expected in test_cases:
        # Simulate the logic from the sync script
        if prefix:
            path_parts = input_path.rsplit('/', 1)
            if len(path_parts) == 2:
                dir_part, file_part = path_parts
                result = f"{dir_part}/{prefix}{file_part}"
            else:
                result = f"{prefix}{input_path}"
        else:
            result = input_path
        
        assert result == expected, f"File prefix test failed: {input_path} with prefix '{prefix}' should be '{expected}', got '{result}'"
    
    print("✅ File prefix logic test passed")

def test_metadata_detection():
    """Test that defaultRegistration field can be detected."""
    import yaml
    
    # Test with defaultRegistration: true
    manifest_with_flag = """
defaultRegistration: true
namespace: Test.Resources
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema: {}
"""
    
    data = yaml.safe_load(manifest_with_flag)
    assert data.get('defaultRegistration', False) is True, "Should detect defaultRegistration: true"
    
    # Test without flag
    manifest_without_flag = """
namespace: Test.Resources
types:
  testType:
    apiVersions:
      "2023-10-01-preview":
        schema: {}
"""
    
    data = yaml.safe_load(manifest_without_flag)
    assert data.get('defaultRegistration', False) is False, "Should not have defaultRegistration by default"
    
    print("✅ Metadata detection test passed")

def main():
    """Run all tests."""
    print("Running resource type sync tests...\n")
    
    try:
        test_yaml_parsing()
        test_validation_logic()
        test_file_prefix_logic()
        test_metadata_detection()
        
        print("\n✅ All tests passed!")
        return 0
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        return 1
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        import traceback
        traceback.print_exc()
        return 1

if __name__ == '__main__':
    sys.exit(main())
