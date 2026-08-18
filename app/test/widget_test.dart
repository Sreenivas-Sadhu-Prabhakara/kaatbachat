import 'package:flutter_test/flutter_test.dart';

import 'package:kaatbachat_app/main.dart';

void main() {
  test('single-length packing = ceiling division', () {
    final p = plan(6, [Piece(2, 10)], 5);
    expect(p.single, true);
    expect(p.stockUnits, 4);
  });

  test('multi-length packing never overflows a stock unit', () {
    final p = plan(6, [Piece(2.3, 3), Piece(1.7, 5), Piece(0.9, 2)], 5);
    expect(p.single, false);
    for (final b in p.bins) {
      expect(b.fold<double>(0, (s, L) => s + L) <= 6 + 1e-9, true);
    }
  });

  testWidgets('renders the cutting-plan title', (tester) async {
    await tester.pumpWidget(const KaatbachatApp());
    expect(find.text('Required pieces'), findsOneWidget);
  });
}
