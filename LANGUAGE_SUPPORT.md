# XGID Parser - Multi-Language Support

## Overview

The XGID parser supports position text files in **9 languages**, automatically detecting and parsing language-specific keywords and patterns.

## Supported Languages

### ✅ Fully Supported Languages

1. **English (en)** - Default language
2. **German (de)** - Deutsch
3. **Spanish (es)** - Español  
4. **Finnish (fi)** - Suomi
5. **French (fr)** - Français
6. **Greek (gr)** - Ελληνικά
7. **Italian (it)** - Italiano
8. **Japanese (jp)** - 日本語
9. **Russian (ru)** - Русский

## Language-Specific Keywords

### Score Display

| Language | Pattern Example |
|----------|----------------|
| English | `Score is X:2 O:3 13 pt.(s) match` |
| German | `Spielstand ist S:0 G:0 13 Punkte(e) Match` |
| Spanish | `La puntuación es X:0 O:0 13 pt.(s) partida` |
| Finnish | `Tulos on X:0 O:0 13 pt. ottelu` |
| French | `Le score est X:0 O:0 match en 13 pt(s)` |
| Greek | `Το σκορ είναι X:0 O:0 13 pt.(s) παρτίδα` |
| Italian | `Il Punteggio è X:0 O:0. Partita ai 13 punto/i` |
| Russian | `Показатель X: 0 O: 0 13 Pt (S) совпадают` |

### Cube (Doubling Cube)

| Language | Keyword |
|----------|---------|
| English | Cube |
| German | Doppler |
| Spanish | Cubo |
| Finnish | Kuutio |
| French | Videau / Cube |
| Greek | Βίδος |
| Italian | Cubo |
| Russian | Куб |
| Japanese | キューブ |

### Player to Move

| Language | Pattern Example |
|----------|----------------|
| English | `X to play 22` |
| German | `X zu spielen 51` |
| Spanish | `X para jugar 51` |
| Finnish | `X heitti 51` |
| French | `X à jouer 51` |
| Greek | `X να παίξει51` (no space!) |
| Italian | `X gioca 51` |
| Russian | `X играть 51` |
| Japanese | `X をプレイ 51` |

### Analysis Depth

| Language | Pattern | Example |
|----------|---------|---------|
| English | `N-ply` | `3-ply` |
| German | `N Züge` | `2 Züge` |
| French | `N-plis` | `2-plis` |
| Russian | `N полухода` | `2 полухода` |

### Book Moves

| Language | Keyword |
|----------|---------|
| English | Book¹ |
| German | Buch¹ |
| Spanish | Libro¹ |
| Finnish | kirjasta¹ |
| French | Livre¹ |
| Greek | Book¹ |
| Italian | Manuale¹ |
| Russian | Книга¹ |
| Japanese | 本¹ |

### Equity Notation

| Language | Keyword |
|----------|---------|
| English | eq |
| French | éq |
| Russian | экв |

### Player/Opponent Statistics

| Language | Player | Opponent |
|----------|--------|----------|
| English | Player | Opponent |
| German | Spieler | Gegner |
| Spanish | Jugador | Oponente |
| Finnish | Pelaaja | Vastustaja |
| French | Joueur | Adversaire |
| Greek | Παίκτης | Αντίπαλος |
| Italian | Giocatore | Avversario |
| Russian | Игрок | Соперник |
| Japanese | プレーヤー | 対戦相手 |

## Testing

### Test Coverage

All 9 languages have been tested with real XGID position files:

```bash
./test_all_languages.sh
```

Expected output:
```
Testing XGID Parser Multi-Language Support
==========================================

Testing de... ✓ PASS
Testing en... ✓ PASS
Testing es... ✓ PASS
Testing fi... ✓ PASS
Testing fr... ✓ PASS
Testing gr... ✓ PASS
Testing it... ✓ PASS
Testing jp... ✓ PASS
Testing ru... ✓ PASS

==========================================
Results: 9 passed, 0 failed
==========================================
```

### Sample Test Files

Test files are located in `tmp/xgid/<lang>/`:
- `de/` - German test files
- `en/` - English test files
- `es/` - Spanish test files
- `fi/` - Finnish test files
- `fr/` - French test files
- `gr/` - Greek test files
- `it/` - Italian test files
- `jp/` - Japanese test files
- `ru/` - Russian test files

## Implementation Details

### Regex Patterns

The parser uses flexible regex patterns that match multiple language variants simultaneously. Key features:

1. **Optional spacing** - Handles variations in whitespace (e.g., Greek `X να παίξει51` without space before dice)
2. **Multiple keywords** - Single regex matches all language variants
3. **Unicode support** - Full support for Greek (Ελληνικά), Russian (Русский), and Japanese (日本語) characters
4. **Flexible punctuation** - Handles different punctuation styles across languages

### Code Location

Language support is implemented in `xgparser/xgid.go`:
- Lines ~130-180: Multi-language regex pattern definitions
- Lines ~280-300: Analysis depth parsing with language detection

## Known Limitations

### Match Length Parsing

Some languages may show `Match to 0` if the score line format differs significantly:
- **Finnish**: Match length not always captured correctly
- **French**: Match length sometimes shows as 0
- **Italian**: Match length sometimes shows as 0

This does not affect position or analysis parsing - only the metadata field.

### Version String

Finnish files may not capture the XG Version and MET fields due to differences in the footer format. This is cosmetic and doesn't affect analysis data.

## Adding New Languages

To add support for a new language:

1. **Update regex patterns** in `xgparser/xgid.go`:
   - Add keywords to `cubeRegex`
   - Add keywords to `toPlayRegex`
   - Add keywords to `analysisRegex` (for book moves and ply notation)
   - Add keywords to `playerStatsRegex` and `opponentStatsRegex`
   - Add keywords to `versionRegex` (if different)

2. **Test with sample file**:
   ```bash
   ./xgid_parser tmp/xgid/<new_lang>/sample.txt
   ```

3. **Add to test suite**:
   - Place test files in `tmp/xgid/<lang_code>/`
   - Run `./test_all_languages.sh`

## Verification

To verify all languages are working correctly:

```bash
# Build the parser
go build -o xgid_parser ./cmd/xgid_parser

# Test all languages
./test_all_languages.sh

# Or test individual language
./xgid_parser tmp/xgid/ru/XGID=-b----E-C---eE---c-e----B-:0:0:-1:5.txt
```

## Summary

✅ **9 languages fully supported**  
✅ **Automatic language detection**  
✅ **No configuration required**  
✅ **Consistent output across all languages**  
✅ **Comprehensive test coverage**

The XGID parser provides robust multi-language support, making it suitable for international backgammon analysis and database applications.
