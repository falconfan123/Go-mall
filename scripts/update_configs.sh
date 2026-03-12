#!/bin/bash

# Update database configuration from jjzzchtt:jjzzchtt to root:fht3825099
# and remove Redis password

# Find all yaml config files
CONFIG_FILES=$(find /Users/fan/go-mall -name "*.yaml" -type f | grep -E "(services|apis)" | grep -v ".prod.yaml" | grep -v "manifests")

for file in $CONFIG_FILES; do
  echo "Updating $file..."

  # Update MySQL datasource
  sed -i '' 's/jjzzchtt:jjzzchtt@tcp/root:fht3825099@tcp/g' "$file"

  # Update Redis password - set Pass to empty
  sed -i '' 's/Pass: jjzzchtt/Pass: ""/g' "$file"
  sed -i '' 's/Pass: jjzzchtt # 如果有密码则填写/Pass: ""/g' "$file"
done

echo "Config update complete!"
