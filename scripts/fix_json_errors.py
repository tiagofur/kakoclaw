import os
import re

def fix_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()

    new_content = re.sub(r'(\s+)json\.NewEncoder\((.*?)\)\.Encode\(', r'\1writeJSONResponse(\2, ', content)

    if content != new_content:
        with open(path, 'w', encoding='utf-8') as f:
            f.write(new_content)
        print(f"Fixed {path}")

def main():
    directory = r"c:\Users\tfurt\source\repos\kakoclaw\pkg\web"
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith(".go"):
                fix_file(os.path.join(root, file))

if __name__ == "__main__":
    main()
