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

for file in logs/*; do
    printf '<h3><a href="%s">%s</a></h3>' "$file" "$file"

    echo Cross posts:
    echo '<pre>'
    tail -n 1000 "$file" | sed -ne '/possible cross-post/,/----/p' | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
    echo '</pre>'

    echo tail:
    echo '<pre>'
    tail "$file" | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
    echo '</pre>'
done

cat << EOF

</body>
</html>
EOF
