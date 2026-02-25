import os

def fix_file(path):
    with open(path, 'r', encoding='utf-8') as f:
        content = f.read()

    # We do a simple string replacement since all calls in this package use the same format.
    # Note: handlers might use 'w' or other variable names but the grep result showed they all use 'w'
    # except maybe we should just replace `_ = json.NewEncoder(w).Encode(` with `writeJSONResponse(w, `
    # and `_ = json.NewEncoder(sw).Encode(` with no replacements needed since we didn't see `sw` in grep.
    # Let's verify by replacing `_ = json.NewEncoder(w).Encode(` -> `writeJSONResponse(w, `
    new_content = content.replace("_ = json.NewEncoder(w).Encode(", "writeJSONResponse(w, ")

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
