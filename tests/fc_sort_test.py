import unittest

import fc_sort

class FcSortTest(unittest.TestCase):

  def testGetStemLen(self):
    self.assertEqual(fc_sort.get_stem_len("/data"), 5)
    self.assertEqual(fc_sort.get_stem_len("/data/system"), 12)
    self.assertEqual(fc_sort.get_stem_len("/data/(system)?"), 6)

  def testIsMeta(self):
    self.assertEqual(fc_sort.is_meta("/data"), False)
    self.assertEqual(fc_sort.is_meta("/data$"), True)
    self.assertEqual(fc_sort.is_meta("\$data"), False)

  def testReadFileContexts(self):
    content = """# comment
/                   u:object_r:rootfs:s0
# another comment
/adb_keys           u:object_r:adb_keys_file:s0
"""
    fcs = fc_sort.read_file_contexts(content.splitlines())
    self.assertEqual(len(fcs), 2)

    self.assertEqual(fcs[0].path, "/")
    self.assertEqual(fcs[0].type, "rootfs")

    self.assertEqual(fcs[1].path, "/adb_keys")
    self.assertEqual(fcs[1].type, "adb_keys_file")

if __name__ == '__main__':
  unittest.main(verbosity=2)

