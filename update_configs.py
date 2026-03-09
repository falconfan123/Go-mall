import os
import glob

def update_file(file_path):
    try:
        with open(file_path, 'r') as f:
            content = f.read()

        original = content

        # Update MySQL datasource
        content = content.replace('jjzzchtt:jjzzchtt@tcp', 'root:fht3825099@tcp')

        # Update Redis password
        content = content.replace('Pass: jjzzchtt', 'Pass: ""')
        content = content.replace('Pass: jjzzchtt # 如果有密码则填写', 'Pass: ""')

        if content != original:
            with open(file_path, 'w') as f:
                f.write(content)
            print(f"Updated: {file_path}")
        else:
            print(f"No changes: {file_path}")
    except Exception as e:
        print(f"Error processing {file_path}: {e}")

def main():
    base_dir = '/Users/fan/go-mall'

    # Find all yaml files in services and apis directories
    patterns = [
        os.path.join(base_dir, 'services', '*', 'etc', '*.yaml'),
        os.path.join(base_dir, 'apis', '*', 'etc', '*.yaml'),
    ]

    for pattern in patterns:
        for file_path in glob.glob(pattern):
            if '.prod.yaml' not in file_path:
                update_file(file_path)

if __name__ == '__main__':
    main()
