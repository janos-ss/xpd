#!/usr/bin/env bash

cross_post_log_start='possible cross-post:$'

cross_posts() {
    awk -v mark="$cross_post_log_start" \
    '$0 ~ mark { print; getline; print; getline; while ($0 ~ /^  of:/) { print; getline; } print "----" }'
}

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
    printf '<h3><a href="%s">%s</a></h3>\n\n' "$file" "$file"

    if tail -n 1000 "$file" | grep -q "$cross_post_log_start"; then
        echo '<pre>'
        tail -n 1000 "$file" | cross_posts | tail -n 20 | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
        echo '</pre>'
    fi

    echo
    echo '<pre>'
    tail "$file" | sed -e 's!htt.*://[^)]*!<a href="&">&</a>!g'
    echo '</pre>'
done

cat << EOF

</body>
</html>
EOF
