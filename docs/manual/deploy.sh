#!/bin/bash
npm run build
rsync -avz --delete build/ nelson@82.25.95.39:/var/www/naphsoft/docs/
echo "¡Documentación desplegada! 📚"
