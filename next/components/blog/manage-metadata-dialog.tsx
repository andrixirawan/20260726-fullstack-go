"use client"

import { useState, useEffect } from "react"
import { Plus, Trash2, Edit, Save, X, Folder, Tag as TagIcon, AlertCircle } from "lucide-react"
import { toast } from "sonner"

import { blogApi, Category, Tag, ApiError } from "@/lib/api"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import { Separator } from "@/components/ui/separator"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"

interface ManageMetadataDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onReload: () => void
}

export function ManageMetadataDialog({ open, onOpenChange, onReload }: ManageMetadataDialogProps) {
  const [activeTab, setActiveTab] = useState<"categories" | "tags">("categories")

  // Data states
  const [categories, setCategories] = useState<Category[]>([])
  const [tags, setTags] = useState<Tag[]>([])

  // Category form states
  const [newCatName, setNewCatName] = useState("")
  const [newCatDesc, setNewCatDesc] = useState("")
  const [editingCatId, setEditingCatId] = useState<string | null>(null)
  const [editingCatName, setEditingCatName] = useState("")
  const [editingCatDesc, setEditingCatDesc] = useState("")

  // Tag form states
  const [newTagName, setNewTagName] = useState("")

  const [loading, setLoading] = useState(false)

  // Fetch all categories and tags
  const loadData = async () => {
    try {
      const [cats, tg] = await Promise.all([blogApi.listCategories(), blogApi.listTags()])
      setCategories(cats)
      setTags(tg)
    } catch {
      toast.error("Failed to load metadata")
    }
  }

  useEffect(() => {
    if (open) {
      loadData()
    }
  }, [open])

  // CATEGORIES Actions
  const handleCreateCategory = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newCatName.trim()) return
    setLoading(true)
    try {
      await blogApi.createCategory({ name: newCatName, description: newCatDesc })
      setNewCatName("")
      setNewCatDesc("")
      toast.success("Category created successfully!")
      loadData()
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to create category")
    } finally {
      setLoading(false)
    }
  }

  const handleUpdateCategory = async (id: string) => {
    if (!editingCatName.trim()) return
    setLoading(true)
    try {
      await blogApi.updateCategory(id, { name: editingCatName, description: editingCatDesc })
      setEditingCatId(null)
      toast.success("Category updated successfully!")
      loadData()
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to update category")
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteCategory = async (id: string, name: string) => {
    if (
      !confirm(
        `Delete category "${name}"? Existing posts using this category will NOT be deleted, but will become uncategorized.`
      )
    ) {
      return
    }
    try {
      await blogApi.deleteCategory(id)
      toast.success("Category deleted")
      loadData()
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete category")
    }
  }

  // TAGS Actions
  const handleCreateTag = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTagName.trim()) return
    setLoading(true)
    try {
      await blogApi.createTag({ name: newTagName })
      setNewTagName("")
      toast.success("Tag created successfully!")
      loadData()
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to create tag")
    } finally {
      setLoading(false)
    }
  }

  const handleDeleteTag = async (id: string, name: string) => {
    if (
      !confirm(
        `Delete tag "#${name}"? Existing posts associated with this tag will NOT be deleted, but will lose this tag.`
      )
    ) {
      return
    }
    try {
      await blogApi.deleteTag(id)
      toast.success("Tag deleted")
      loadData()
      onReload()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Failed to delete tag")
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg w-full max-h-[85vh] flex flex-col p-6 overflow-hidden">
        <DialogHeader>
          <DialogTitle>Manage Blog Metadata</DialogTitle>
          <DialogDescription>
            Create, update, or delete blog categories and tags.
          </DialogDescription>
        </DialogHeader>

        {/* Custom Tabs */}
        <div className="flex border-b border-border mt-2">
          <button
            onClick={() => setActiveTab("categories")}
            className={`flex items-center gap-1.5 px-4 py-2 border-b-2 text-sm font-medium transition-colors ${
              activeTab === "categories"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <Folder className="h-4 w-4" />
            Categories ({categories.length})
          </button>
          <button
            onClick={() => setActiveTab("tags")}
            className={`flex items-center gap-1.5 px-4 py-2 border-b-2 text-sm font-medium transition-colors ${
              activeTab === "tags"
                ? "border-primary text-primary"
                : "border-transparent text-muted-foreground hover:text-foreground"
            }`}
          >
            <TagIcon className="h-4 w-4" />
            Tags ({tags.length})
          </button>
        </div>

        {/* Tab Content */}
        <div className="flex-1 overflow-y-auto pt-4 space-y-6 pr-1">
          {activeTab === "categories" ? (
            <div className="space-y-6">
              {/* Category creation form */}
              <form onSubmit={handleCreateCategory} className="space-y-3 bg-muted/40 p-3 rounded-xl border border-border/50">
                <h4 className="font-semibold text-xs text-muted-foreground uppercase tracking-wider">
                  Create New Category
                </h4>
                <div className="space-y-1">
                  <Label htmlFor="cat-name-input" className="text-xs">Name</Label>
                  <Input
                    id="cat-name-input"
                    placeholder="e.g. Technology"
                    value={newCatName}
                    onChange={(e) => setNewCatName(e.target.value)}
                    className="h-8 text-xs"
                    required
                  />
                </div>
                <div className="space-y-1">
                  <Label htmlFor="cat-desc-input" className="text-xs">Description</Label>
                  <Textarea
                    id="cat-desc-input"
                    placeholder="Brief description..."
                    value={newCatDesc}
                    onChange={(e) => setNewCatDesc(e.target.value)}
                    rows={2}
                    className="text-xs min-h-[50px] resize-none"
                  />
                </div>
                <Button type="submit" size="sm" className="w-full gap-1 h-8 text-xs" disabled={loading}>
                  <Plus className="h-3.5 w-3.5" />
                  Add Category
                </Button>
              </form>

              <Separator />

              {/* Category List */}
              <div className="space-y-3">
                <h4 className="font-semibold text-xs text-muted-foreground uppercase tracking-wider">
                  All Categories
                </h4>
                <div className="space-y-2">
                  {categories.map((cat) => {
                    const isEditing = editingCatId === cat.id
                    return (
                      <div
                        key={cat.id}
                        className="flex flex-col gap-2 p-3 border border-border/40 bg-card rounded-lg hover:border-border transition-colors group"
                      >
                        {isEditing ? (
                          <div className="space-y-2 w-full">
                            <Input
                              value={editingCatName}
                              onChange={(e) => setEditingCatName(e.target.value)}
                              className="h-8 text-xs"
                              placeholder="Category Name"
                              autoFocus
                            />
                            <Textarea
                              value={editingCatDesc}
                              onChange={(e) => setEditingCatDesc(e.target.value)}
                              rows={2}
                              className="text-xs resize-none"
                              placeholder="Description"
                            />
                            <div className="flex gap-2">
                              <Button
                                size="xs"
                                className="gap-1"
                                onClick={() => handleUpdateCategory(cat.id)}
                                disabled={loading}
                              >
                                <Save className="h-3 w-3" />
                                Save
                              </Button>
                              <Button
                                size="xs"
                                variant="ghost"
                                onClick={() => setEditingCatId(null)}
                              >
                                <X className="h-3 w-3" />
                                Cancel
                              </Button>
                            </div>
                          </div>
                        ) : (
                          <div className="flex items-start justify-between gap-3 w-full">
                            <div className="space-y-0.5">
                              <span className="font-medium text-xs text-foreground flex items-center gap-1.5">
                                {cat.name}
                                <span className="text-[10px] text-muted-foreground bg-muted px-1.5 py-0.2 rounded-full font-mono">
                                  /{cat.slug}
                                </span>
                              </span>
                              {cat.description && (
                                <p className="text-xs text-muted-foreground/80 leading-normal">
                                  {cat.description}
                                </p>
                              )}
                            </div>
                            <div className="flex gap-1 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                              <Button
                                size="icon-xs"
                                variant="ghost"
                                onClick={() => {
                                  setEditingCatId(cat.id)
                                  setEditingCatName(cat.name)
                                  setEditingCatDesc(cat.description)
                                }}
                                title="Edit Category"
                              >
                                <Edit className="h-3 w-3 text-muted-foreground hover:text-foreground" />
                              </Button>
                              <Button
                                size="icon-xs"
                                variant="ghost"
                                onClick={() => handleDeleteCategory(cat.id, cat.name)}
                                title="Delete Category"
                              >
                                <Trash2 className="h-3 w-3 text-destructive hover:text-destructive" />
                              </Button>
                            </div>
                          </div>
                        )}
                      </div>
                    )
                  })}
                  {categories.length === 0 && (
                    <p className="text-xs text-muted-foreground text-center py-4">No categories created yet.</p>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="space-y-6">
              {/* Tag creation form */}
              <form onSubmit={handleCreateTag} className="flex gap-2 bg-muted/40 p-3 rounded-xl border border-border/50">
                <div className="flex-1 space-y-1">
                  <Input
                    placeholder="Enter new tag name (e.g. go, react)..."
                    value={newTagName}
                    onChange={(e) => setNewTagName(e.target.value)}
                    className="h-8 text-xs"
                    required
                  />
                </div>
                <Button type="submit" size="sm" className="gap-1 h-8 text-xs" disabled={loading}>
                  <Plus className="h-3.5 w-3.5" />
                  Add Tag
                </Button>
              </form>

              <Separator />

              {/* Tag List */}
              <div className="space-y-3">
                <h4 className="font-semibold text-xs text-muted-foreground uppercase tracking-wider">
                  All Tags
                </h4>
                <div className="flex flex-wrap gap-2">
                  {tags.map((tag) => (
                    <div
                      key={tag.id}
                      className="inline-flex items-center gap-1 bg-primary/5 text-primary text-xs pl-2.5 pr-1.5 py-1 rounded-full border border-primary/10 hover:border-primary/30 transition-colors group"
                    >
                      <span>#{tag.name}</span>
                      <button
                        type="button"
                        onClick={() => handleDeleteTag(tag.id, tag.name)}
                        className="text-muted-foreground hover:text-destructive p-0.5 rounded-full hover:bg-destructive/15 transition-colors cursor-pointer"
                        title="Delete Tag"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </div>
                  ))}
                  {tags.length === 0 && (
                    <p className="text-xs text-muted-foreground text-center py-4 w-full">No tags created yet.</p>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer info warning */}
        <div className="mt-4 pt-3 border-t border-border/50 flex gap-2 items-start text-xs text-muted-foreground/80 bg-amber-500/5 p-2 rounded-lg border border-amber-500/10">
          <AlertCircle className="h-4 w-4 text-amber-500 shrink-0 mt-0.5" />
          <p>
            Deleting a category sets related posts to <em>uncategorized</em>. Deleting a tag removes the link without deleting the posts themselves.
          </p>
        </div>
      </DialogContent>
    </Dialog>
  )
}
