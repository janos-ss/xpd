#!/usr/bin/env bash

echo Content-type: text/html
echo "Refresh: 300;$REQUEST_URI"
echo

title="Summary ($(date))"

cat << EOF
<!doctype html>
<html>
<head>
<title>$title</title>
</head>
<body>

<h1>$title</h1>

EOF

if [[ "$QUERY_STRING" =~ ^[a-z.]+$ ]]; then
    pattern="$QUERY_STRING"
else
    pattern=.
fi

for file in $(ls logs | grep $pattern); do
    file=logs/$file
    printf '<h3><a href="%s">%s</a></h3>' "$file" "$file"

    if tail -n 1000 "$file" | grep -qF 'possible cross-post'; then
        echo '<pre>'
        tail -n 1000 "$file" | sed -ne '/possible cross-post/,/----/p' | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
        echo '</pre>'
    fi

    echo '<pre>'
    tail "$file" | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
    echo '</pre>'
done

cat << EOF

</body>
</html>
EOF
