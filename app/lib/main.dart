import 'package:flutter/material.dart';

void main() => runApp(const KaatbachatApp());

/// Kaatbachat — cut-to-length planner. Packs several required lengths into the
/// fewest stock units (first-fit-decreasing), mirroring the Go engine.
class KaatbachatApp extends StatelessWidget {
  const KaatbachatApp({super.key});
  @override
  Widget build(BuildContext context) => MaterialApp(
        title: 'Kaatbachat',
        debugShowCheckedModeBanner: false,
        theme: ThemeData(colorSchemeSeed: const Color(0xFF444B57), useMaterial3: true),
        home: const HomePage(),
      );
}

class Piece {
  double length;
  int count;
  Piece(this.length, this.count);
}

class Plan {
  final int stockUnits, cuts;
  final double offcut, labourCost;
  final List<List<double>> bins;
  final bool single;
  const Plan(this.stockUnits, this.cuts, this.offcut, this.labourCost, this.bins, this.single);
}

/// plan mirrors backend/cost.go: single-length short-circuit, else FFD packing.
Plan plan(double stock, List<Piece> pieces, double cutLabour) {
  final distinct = pieces.map((p) => p.length).toSet();
  int total = pieces.fold(0, (s, p) => s + p.count);
  List<List<double>> bins = [];
  bool single = false;

  if (distinct.length == 1 && pieces.isNotEmpty) {
    single = true;
    final len = pieces.first.length;
    var per = (stock / len).floor();
    if (per < 1) per = 1;
    var remaining = total;
    while (remaining > 0) {
      final n = remaining < per ? remaining : per;
      bins.add(List.filled(n, len));
      remaining -= n;
    }
  } else {
    final lengths = <double>[];
    for (final p in pieces) {
      for (var i = 0; i < p.count; i++) lengths.add(p.length);
    }
    lengths.sort((a, b) => b.compareTo(a));
    final remain = <double>[];
    for (final L in lengths) {
      var placed = false;
      for (var i = 0; i < bins.length; i++) {
        if (remain[i] >= L) {
          bins[i].add(L);
          remain[i] -= L;
          placed = true;
          break;
        }
      }
      if (!placed) {
        bins.add([L]);
        remain.add(stock - L);
      }
    }
  }

  double offcut = 0;
  int cuts = 0;
  for (final b in bins) {
    double used = 0;
    for (final L in b) {
      used += L;
      cuts++;
    }
    offcut += stock - used;
  }
  return Plan(bins.length, cuts, offcut, cuts * cutLabour, bins, single);
}

class HomePage extends StatefulWidget {
  const HomePage({super.key});
  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  final _stock = TextEditingController(text: '6');
  final _labour = TextEditingController(text: '5');
  final _pieces = <Piece>[Piece(2.3, 3), Piece(1.7, 5), Piece(0.9, 2)];
  Plan? _plan;

  void _calc() => setState(() => _plan = plan(
        double.tryParse(_stock.text.trim()) ?? 0, _pieces,
        double.tryParse(_labour.text.trim()) ?? 0,
      ));

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Kaatbachat · cutting plan'),
        backgroundColor: Theme.of(context).colorScheme.primaryContainer,
      ),
      body: ListView(padding: const EdgeInsets.all(16), children: [
        Row(children: [
          Expanded(child: _num(_stock, 'Stock length (m)')),
          const SizedBox(width: 12),
          Expanded(child: _num(_labour, 'Cut labour ₹/cut')),
        ]),
        const SizedBox(height: 8),
        const Text('Required pieces', style: TextStyle(fontWeight: FontWeight.w600)),
        for (var i = 0; i < _pieces.length; i++) _pieceRow(i),
        TextButton.icon(
          onPressed: () => setState(() => _pieces.add(Piece(1, 1))),
          icon: const Icon(Icons.add), label: const Text('Add length'),
        ),
        FilledButton.icon(onPressed: _calc, icon: const Icon(Icons.content_cut), label: const Text('Plan the cuts')),
        const SizedBox(height: 20),
        if (_plan != null) _result(),
      ]),
    );
  }

  Widget _num(TextEditingController c, String label) => TextField(
        controller: c,
        keyboardType: const TextInputType.numberWithOptions(decimal: true),
        decoration: InputDecoration(labelText: label, border: const OutlineInputBorder()),
        onChanged: (_) => _calc(),
      );

  Widget _pieceRow(int i) => Padding(
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Row(children: [
          Expanded(child: TextField(
            decoration: const InputDecoration(labelText: 'Length m', border: OutlineInputBorder()),
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            controller: TextEditingController(text: _pieces[i].length.toString()),
            onChanged: (v) { _pieces[i].length = double.tryParse(v) ?? 0; },
          )),
          const SizedBox(width: 8),
          SizedBox(width: 90, child: TextField(
            decoration: const InputDecoration(labelText: 'Count', border: OutlineInputBorder()),
            keyboardType: TextInputType.number,
            controller: TextEditingController(text: _pieces[i].count.toString()),
            onChanged: (v) { _pieces[i].count = int.tryParse(v) ?? 0; },
          )),
          IconButton(onPressed: () => setState(() => _pieces.removeAt(i)), icon: const Icon(Icons.close)),
        ]),
      );

  Widget _result() {
    final p = _plan!;
    return Card(
      color: Theme.of(context).colorScheme.primaryContainer,
      child: Padding(padding: const EdgeInsets.all(16),
        child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
          Text('${p.stockUnits} stock unit${p.stockUnits == 1 ? '' : 's'} needed',
              style: const TextStyle(fontSize: 24, fontWeight: FontWeight.bold)),
          Text('Offcut ${p.offcut.toStringAsFixed(2)} m · ${p.cuts} cuts · labour ₹${p.labourCost.toStringAsFixed(2)}'),
          const Divider(),
          for (var i = 0; i < p.bins.length; i++)
            Text('Stock ${i + 1}: ${p.bins[i].map((L) => L.toStringAsFixed(2)).join(' + ')}'),
        ])),
    );
  }
}
