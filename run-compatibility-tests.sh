#!/bin/sh

# Script de teste de compatibilidade entre a versão C e Go do figlet
# Testa todas as fonts disponíveis com uma palavra de teste
#
# Uso: ./run-compatibility-tests.sh [palavra_de_teste]
# Exemplo: ./run-compatibility-tests.sh "HELLO"

LC_ALL=POSIX
export LC_ALL

TESTWORD="${1:-TEST}"
OUTPUT_C=`mktemp`
OUTPUT_GO=`mktemp`
LOGFILE=compatibility-test.log
FIGLET_C=figlet
FIGLET_GO=${FIGLET_BIN:-./figlet-bin}
FONTDIR="fonts"

# Contadores
total=0
passed=0
failed=0
skipped=0

# Limpar log anterior
rm -f "$LOGFILE"

echo "=========================================" | tee -a "$LOGFILE"
echo "Teste de Compatibilidade Figlet C vs Go" | tee -a "$LOGFILE"
echo "=========================================" | tee -a "$LOGFILE"
echo "Palavra de teste: $TESTWORD" | tee -a "$LOGFILE"
echo "Versão C: $FIGLET_C" | tee -a "$LOGFILE"
echo "Versão Go: $FIGLET_GO" | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

# Verificar se os executáveis existem
if ! command -v "$FIGLET_C" > /dev/null 2>&1; then
    echo "ERRO: $FIGLET_C não encontrado no sistema" | tee -a "$LOGFILE"
    exit 1
fi

if [ ! -x "$FIGLET_GO" ]; then
    echo "ERRO: $FIGLET_GO não encontrado ou não é executável" | tee -a "$LOGFILE"
    echo "Compilando a versão Go..." | tee -a "$LOGFILE"
    if ! go build -o "$FIGLET_GO" figlet.go; then
        echo "ERRO: Falha ao compilar a versão Go" | tee -a "$LOGFILE"
        exit 1
    fi
fi

# Listar todas as fonts .flf
FONTS=$(ls "$FONTDIR"/*.flf 2>/dev/null | xargs -n1 basename | sed 's/\.flf$//')

if [ -z "$FONTS" ]; then
    echo "ERRO: Nenhuma font encontrada em $FONTDIR" | tee -a "$LOGFILE"
    exit 1
fi

echo "Fonts encontradas: $(echo "$FONTS" | wc -l)" | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

# Testar cada font
for font in $FONTS; do
    total=$((total + 1))
    printf "Testando font: %-30s " "$font" | tee -a "$LOGFILE"
    
    # Gerar com versão C (usando opções padrão)
    if ! echo "$TESTWORD" | "$FIGLET_C" -f "$font" -d "$FONTDIR" -w 80 > "$OUTPUT_C" 2>/dev/null; then
        echo "SKIP (erro na versão C)" | tee -a "$LOGFILE"
        skipped=$((skipped + 1))
        continue
    fi
    
    # Gerar com versão Go (usando as mesmas opções)
    if ! echo "$TESTWORD" | "$FIGLET_GO" -f "$font" -d "$FONTDIR" -w 80 > "$OUTPUT_GO" 2>/dev/null; then
        echo "SKIP (erro na versão Go)" | tee -a "$LOGFILE"
        skipped=$((skipped + 1))
        continue
    fi
    
    # Normalizar saídas (remover espaços em branco no final das linhas)
    sed 's/[[:space:]]*$//' "$OUTPUT_C" > "${OUTPUT_C}.norm"
    sed 's/[[:space:]]*$//' "$OUTPUT_GO" > "${OUTPUT_GO}.norm"
    
    # Comparar resultados normalizados
    if cmp -s "${OUTPUT_C}.norm" "${OUTPUT_GO}.norm"; then
        echo "PASS" | tee -a "$LOGFILE"
        passed=$((passed + 1))
    else
        echo "FAIL" | tee -a "$LOGFILE"
        failed=$((failed + 1))
        echo "  Diferenças encontradas para a font: $font" >> "$LOGFILE"
        echo "  --- Versão C (primeiras 10 linhas) ---" >> "$LOGFILE"
        head -10 "$OUTPUT_C" >> "$LOGFILE"
        echo "  --- Versão Go (primeiras 10 linhas) ---" >> "$LOGFILE"
        head -10 "$OUTPUT_GO" >> "$LOGFILE"
        echo "  --- Diff (normalizado) ---" >> "$LOGFILE"
        diff "${OUTPUT_C}.norm" "${OUTPUT_GO}.norm" >> "$LOGFILE" 2>&1 || true
        echo "" >> "$LOGFILE"
    fi
done

# Limpar arquivos temporários
rm -f "$OUTPUT_C" "$OUTPUT_GO" "${OUTPUT_C}.norm" "${OUTPUT_GO}.norm"

# Resumo
echo "" | tee -a "$LOGFILE"
echo "=========================================" | tee -a "$LOGFILE"
echo "Resumo do Teste" | tee -a "$LOGFILE"
echo "=========================================" | tee -a "$LOGFILE"
echo "Total de fonts testadas: $total" | tee -a "$LOGFILE"
echo "Testes passados: $passed" | tee -a "$LOGFILE"
echo "Testes falhados: $failed" | tee -a "$LOGFILE"
echo "Testes ignorados: $skipped" | tee -a "$LOGFILE"
echo "" | tee -a "$LOGFILE"

if [ $failed -eq 0 ]; then
    echo "✓ Todos os testes passaram!" | tee -a "$LOGFILE"
    exit 0
else
    echo "✗ $failed teste(s) falharam. Veja $LOGFILE para detalhes." | tee -a "$LOGFILE"
    exit 1
fi
